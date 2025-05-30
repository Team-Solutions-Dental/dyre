package transpiler

import (
	"errors"
	"fmt"

	"github.com/vamuscari/dyre/ast"
	"github.com/vamuscari/dyre/endpoint"
	"github.com/vamuscari/dyre/lexer"
	"github.com/vamuscari/dyre/object"
	"github.com/vamuscari/dyre/parser"
	"github.com/vamuscari/dyre/sql"
	"github.com/vamuscari/dyre/utils"
)

// Intermediate Representation
type IR struct {
	endpoint     *endpoint.Endpoint
	currentField *endpoint.Field
	ast          *ast.QueryStatements
	sql          *sql.Query
	joins        []*joinIR
}

func New(query string, endpoint *endpoint.Endpoint) *IR {
	l := lexer.New(query)
	p := parser.New(l)
	q := p.ParseQuery()
	var ir IR = IR{endpoint: endpoint, ast: q, sql: &sql.Query{}}
	return &ir
}

func (ir *IR) EvaluateQuery() (string, error) {
	result := ir.evalTable()
	if isError(result) {
		return "", errors.New(result.String())
	}
	return ir.sql.ConstructQuery(), nil
}

// TODO: reorder columns for select if found
func (ir *IR) evalTable() object.Object {

	if ir.endpoint.SchemaName != "" {
		ir.sql.From = ir.endpoint.SchemaName + "." + ir.endpoint.TableName
	} else {
		ir.sql.From = ir.endpoint.TableName
	}

	for _, js := range ir.joins {
		result := js.child_ir.evalTable()
		if isError(result) {
			return result
		}
	}

	result := eval(ir.ast, ir)
	if isError(result) {
		return result
	}

	return nil
}

func eval(node ast.Node, ir *IR) object.Object {
	switch node := node.(type) {
	case *ast.QueryStatements:
		return evalQueryStatements(node, ir)
	case *ast.ColumnStatement:
		return evalColumnStatement(node, ir)
	case *ast.BlockStatement:
		return evalBlockStatement(node, ir)
	case *ast.ExpressionStatement:
		return evalExpressionStatement(node, ir)
	case *ast.IntegerLiteral:
		return &object.Integer{Value: node.Value}
	case *ast.NullLiteral:
		return &object.Null{}
	case *ast.StringLiteral:
		return &object.String{Value: node.Value}
	case *ast.Boolean:
		return &object.Boolean{Value: node.Value}
	// case *ast.Identifier:
	// 	return "Identifier"
	case *ast.PrefixExpression:
		right := eval(node.Right, ir)
		if isError(right) {
			return right
		}
		return evalPrefixExpression(node.Operator, right)
	case *ast.InfixExpression:
		left := eval(node.Left, ir)
		if isError(left) {
			return left
		}
		right := eval(node.Right, ir)
		if isError(right) {
			return right
		}
		return evalInfixExpression(node.Operator, left, right)
	case *ast.ColumnCall:
		return evalColumnCall(node, ir)
	case *ast.CallExpression:
		return evalCallExpression(node.Function.TokenLiteral(), node.Arguments, ir)
	default:
		return newError(fmt.Sprintf("Unknown Evaluation Type: %T", node))
	}
}

func evalQueryStatements(node *ast.QueryStatements, ir *IR) object.Object {
	var result object.Object

	for _, statement := range node.Statements {
		result = eval(statement, ir)

		switch result := result.(type) {
		case *object.Error:
			return result
		}
	}
	return result
}

func evalColumnStatement(node *ast.ColumnStatement, ir *IR) object.Object {
	if !utils.Array_Contains(ir.endpoint.FieldNames, node.TokenLiteral()) {
		return newError(fmt.Sprintf("Requested column %s not found for %s", node.TokenLiteral(), ir.endpoint.TableName))
	}

	column := ir.endpoint.Fields[node.TokenLiteral()]
	ir.currentField = &column

	// selectStatementLoc := ir.selectStatementLocation(ir.currentField.Name)
	// if selectStatementLoc >= 0 {
	// 	newSelectStatement := []*sql.selectStatement{}
	// 	for i, ss := range ir.sql.selectStatements {
	// 		if i != selectStatementLoc {
	// 			newSelectStatement = append(newSelectStatement, ss)
	// 		}
	// 	}
	//
	// 	newSelectStatement = append(newSelectStatement, ir.sql.selectStatements[selectStatementLoc])
	// 	ir.sql.selectStatements = newSelectStatement
	//
	// } else {
	// 	selects := &selectStatement{fieldName: &ir.currentField.Name, tableName: &ir.endpoint.TableName}
	// 	ir.selectStatements = append(ir.selectStatements, selects)
	// }

	if node.Expressions != nil {
		err := eval(node.Expressions, ir)
		if err != nil {
			return err
		}
	}

	return nil
}

func evalBlockStatement(block *ast.BlockStatement, ir *IR) object.Object {
	var result object.Object

	for _, statement := range block.Statements {
		result = eval(statement, ir)

		if result != nil {
			rt := result.Type()
			if rt == object.ERROR_OBJ {
				return result
			}
		}
	}

	return result
}

func evalExpressionStatement(stmnt *ast.ExpressionStatement, ir *IR) object.Object {
	evaluated := eval(stmnt.Expression, ir)
	if isError(evaluated) {
		return evaluated
	}

	if evaluated.Type() == object.BOOLEAN_OBJ {
		ir.sql.WhereStatements = append(ir.sql.WhereStatements, evaluated.String())
	}

	return nil

}

func evalPrefixExpression(operator string, right object.Object) object.Object {
	switch operator {
	case "!":
		return evalBangOperatorExpression(right)
	case "-":
		return evalMinusPrefixOperatorExpression(right)
	default:
		return newError("unknown operator: %s%s", operator, right.Type())
	}
}

func evalBangOperatorExpression(right object.Object) object.Object {
	switch {
	case right.Type() == object.BOOLEAN_OBJ:
		return &object.BooleanExpression{Value: fmt.Sprintf("!%s", right.String())}
	case right.Type() == object.EXPRESSION_OBJ:
		return &object.BooleanExpression{Value: fmt.Sprintf("!%s", right.String())}
	default:
		return newError("Invalid Bang Operator Expression " + right.String())
	}
}

func evalMinusPrefixOperatorExpression(right object.Object) object.Object {
	switch {
	case right.Type() == object.INTEGER_OBJ:
		return &object.Expression{ExpressionType: object.INTEGER_OBJ, Value: fmt.Sprintf("-%s", right.String())}
	case right.Type() == object.EXPRESSION_OBJ:
		return &object.Expression{ExpressionType: object.INTEGER_OBJ, Value: fmt.Sprintf("-%s", right.String())}
	default:
		return newError("Invalid Minus Prefix Operator Expression" + right.String())
	}
}

func evalInfixExpression(operator string, left, right object.Object) object.Object {
	switch {
	case left.Type() != right.Type():
		return newError("type mismatch: %s %s %s", left.Type(), operator, right.Type())
	case operator == "==":
		return &object.Expression{ExpressionType: object.BOOLEAN_OBJ,
			Value: fmt.Sprintf("(%s = %s)", left.String(), right.String())}
	case operator == "!=":
		return &object.Expression{ExpressionType: object.BOOLEAN_OBJ,
			Value: fmt.Sprintf("(%s != %s)", left.String(), right.String())}
	case operator == ">":
		return &object.Expression{ExpressionType: object.BOOLEAN_OBJ,
			Value: fmt.Sprintf("(%s > %s)", left.String(), right.String())}
	case operator == "<":
		return &object.Expression{ExpressionType: object.BOOLEAN_OBJ,
			Value: fmt.Sprintf("(%s < %s)", left.String(), right.String())}
	case operator == ">=":
		return &object.Expression{ExpressionType: object.BOOLEAN_OBJ,
			Value: fmt.Sprintf("(%s >= %s)", left.String(), right.String())}
	case operator == "<=":
		return &object.Expression{ExpressionType: object.BOOLEAN_OBJ,
			Value: fmt.Sprintf("(%s >= %s)", left.String(), right.String())}
	case operator == "AND":
		return &object.Expression{ExpressionType: object.BOOLEAN_OBJ,
			Value: fmt.Sprintf("(%s AND %s)", left.String(), right.String())}
	case operator == "OR":
		return &object.Expression{ExpressionType: object.BOOLEAN_OBJ,
			Value: fmt.Sprintf("(%s OR %s)", left.String(), right.String())}
	default:
		return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
	}
}

func evalColumnCall(node ast.Node, ir *IR) object.Object {

	if node.TokenLiteral() != "@" {
		return newError(fmt.Sprintf("Invalid Token Literal. got=%s, want=%s", node.TokenLiteral(), "@"))
	}

	return &object.FieldCall{FieldType: ir.currentField.Type(),
		Value: fmt.Sprintf("%s.[%s]", ir.endpoint.TableName, ir.currentField.Name)}
}

func newError(format string, a ...interface{}) *object.Error {
	return &object.Error{Message: fmt.Sprintf(format, a...)}
}

func isError(obj object.Object) bool {
	if obj != nil {
		return obj.Type() == object.ERROR_OBJ
	}
	return false
}

func evalExpressions(
	exps []ast.Expression,
	ir *IR,
) []object.Object {
	var results []object.Object

	for _, e := range exps {
		evaluated := eval(e, ir)
		results = append(results, evaluated)
	}

	return results
}

func evalCallExpression(function string, exps []ast.Expression, ir *IR) object.Object {
	args := evalExpressions(exps, ir)

	return builtins[function](ir, args...)
}

func (ir *IR) INNERJOIN(req string) *joinType {
	join := &joinType{joinType: "INNER", parentIR: ir, name: req}

	return join

}

func (ir *IR) LEFTJOIN(req string) *joinType {
	join := &joinType{joinType: "LEFT", parentIR: ir, name: req}

	return join
}

func (ir *IR) FieldNames() []string {
	return ir.sql.SelectNameList()
}

func (ir *IR) LIMIT(input int) *IR {
	ir.sql.Limit = &input
	return ir
}
