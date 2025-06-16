package transpiler

import (
	"errors"
	"fmt"
	"strings"

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
	endpoint               *endpoint.Endpoint
	currentField           *endpoint.Field
	currentSelectStatement *sql.SelectStatement
	ast                    *ast.RequestStatements
	sql                    *sql.Query
	joins                  []*joinIR
	error                  error
}

type PrimaryIR struct {
	IR
	orderByAST *ast.RequestStatements
}

type SubIR struct {
	IR
}

// Entry into transpiler package
// Create a new query
func New(query string, endpoint *endpoint.Endpoint) (*PrimaryIR, error) {
	if endpoint == nil {
		return nil, errors.New("No end point provided for query: " + query)
	}
	q, err := parse(query)
	var ir PrimaryIR = PrimaryIR{IR: IR{endpoint: endpoint,
		ast:   q,
		error: err,
		sql:   &sql.Query{Depth: 0}}}
	return &ir, err
}

// Sub IR is excludes methods unique to top level query
func newSubIR(query string, endpoint *endpoint.Endpoint) (*SubIR, error) {
	if endpoint == nil {
		return nil, errors.New("No end point provided for query: " + query)
	}
	q, err := parse(query)
	var ir SubIR = SubIR{IR: IR{endpoint: endpoint,
		ast:   q,
		error: err,
		sql:   &sql.Query{Depth: 0}}}
	return &ir, err
}

func parse(req string) (*ast.RequestStatements, error) {
	l := lexer.New(req)
	p := parser.New(l)
	q := p.ParseQuery()
	errs := p.Errors()
	if len(errs) > 0 {
		var sb strings.Builder
		sb.WriteString("Parser Errors:")
		for _, pe := range errs {
			sb.WriteString(pe)
		}
		return q, errors.New(sb.String())
	}

	return q, nil
}

func (pir *PrimaryIR) EvaluateQuery() (string, error) {
	if pir.error != nil {
		return "", pir.error
	}

	result := pir.evalTable()

	if isError(result) {
		pir.error = errors.New(result.String())
		return "", pir.error
	}

	if pir.orderByAST != nil {
		evalOrderBy(pir.orderByAST, &pir.IR)
	}

	return pir.sql.ConstructQuery(), nil
}

// Return a list of names for headers
// Run Evaluate Query First!
func (pir *PrimaryIR) FieldNames() []string {
	return pir.sql.SelectNameList()
}

func (ir *IR) evalTable() object.Object {
	if ir.endpoint.SchemaName != "" {
		ir.sql.From = ir.endpoint.SchemaName + "." + ir.endpoint.TableName
	} else {
		ir.sql.From = ir.endpoint.TableName
	}

	ir.sql.TableName = ir.endpoint.TableName

	for _, js := range ir.joins {
		result := js.childIR.evalTable()
		if isError(result) {
			return result
		}
	}

	result := eval(ir.ast, ir)
	if isError(result) {
		return result
	}

	for _, j := range ir.joins {
		for _, ss := range j.childIR.sql.SelectStatements {
			if ss.Name() == j.childOn {
				continue
			}
			joinedSelect := &sql.SelectStatement{
				FieldName: ss.FieldName,
				TableName: &j.alias,
				Exclude:   ss.Exclude,
			}
			ir.sql.SelectStatements = append(ir.sql.SelectStatements, joinedSelect)
		}

		err := j.Check()
		if err != nil {
			return newError("%s", err.Error())
		}
	}

	return nil
}

func eval(node ast.Node, ir *IR) object.Object {
	switch node := node.(type) {
	case *ast.RequestStatements:
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
		return newError("Unknown Evaluation Type: %T", node)
	}
}

func evalQueryStatements(node *ast.RequestStatements, ir *IR) object.Object {
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
		return newError("Requested column %s not found for %s", node.TokenLiteral(), ir.endpoint.TableName)
	}

	column := ir.endpoint.Fields[node.TokenLiteral()]
	ir.currentField = &column

	selectStatementLoc := ir.sql.SelectStatementLocation(ir.currentField.Name)
	if selectStatementLoc >= 0 {
		newSelectStatementList := []*sql.SelectStatement{}
		for i, ss := range ir.sql.SelectStatements {
			if i != selectStatementLoc {
				newSelectStatementList = append(newSelectStatementList, ss)
			}
		}

		newSelectStatementList = append(newSelectStatementList, ir.sql.SelectStatements[selectStatementLoc])
		ir.currentSelectStatement = ir.sql.SelectStatements[selectStatementLoc]
		ir.sql.SelectStatements = newSelectStatementList

	} else {
		selects := &sql.SelectStatement{FieldName: &ir.currentField.Name, TableName: &ir.endpoint.TableName, Exclude: false}
		ir.currentSelectStatement = selects
		ir.sql.SelectStatements = append(ir.sql.SelectStatements, selects)
	}

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
		if statement != nil {
			result = eval(statement, ir)

			if isError(result) {
				return result
			}
		}
	}

	return result
}

func evalExpressionStatement(stmnt *ast.ExpressionStatement, ir *IR) object.Object {

	if stmnt.Expression == nil {
		return nil
	}

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
		return &object.Expression{
			ExpressionType: object.BOOLEAN_OBJ,
			Value:          fmt.Sprintf("!%s", right.String())}
	default:
		return newError("Invalid Bang Operator Expression %s", right.String())
	}
}

func evalMinusPrefixOperatorExpression(right object.Object) object.Object {
	switch {
	case right.Type() == object.INTEGER_OBJ:
		return &object.Expression{
			ExpressionType: object.INTEGER_OBJ,
			Value:          fmt.Sprintf("-%s", right.String())}
	case right.Type() == object.EXPRESSION_OBJ:
		return &object.Expression{
			ExpressionType: object.INTEGER_OBJ,
			Value:          fmt.Sprintf("-%s", right.String())}
	default:
		return newError("Invalid Minus Prefix Operator Expression %s", right.String())
	}
}

func evalInfixExpression(operator string, left, right object.Object) object.Object {
	switch {
	case left.Type() == object.NULL_OBJ:
		return evalInfixNullExpression(operator, right, left)
	case right.Type() == object.NULL_OBJ:
		return evalInfixNullExpression(operator, left, right)
	// case left.Type() == object.NULL_OBJ || right.Type() == object.NULL_OBJ:
	// 	return newError("type mismatch: %s %s %s", left.Type(), operator, right.Type())
	case operator == "==":
		return &object.Expression{
			ExpressionType: object.BOOLEAN_OBJ,
			Value:          fmt.Sprintf("(%s = %s)", left.String(), right.String())}
	case operator == "!=":
		return &object.Expression{
			ExpressionType: object.BOOLEAN_OBJ,
			Value:          fmt.Sprintf("(%s != %s)", left.String(), right.String())}
	case operator == ">":
		return &object.Expression{
			ExpressionType: object.BOOLEAN_OBJ,
			Value:          fmt.Sprintf("(%s > %s)", left.String(), right.String())}
	case operator == "<":
		return &object.Expression{
			ExpressionType: object.BOOLEAN_OBJ,
			Value:          fmt.Sprintf("(%s < %s)", left.String(), right.String())}
	case operator == ">=":
		return &object.Expression{
			ExpressionType: object.BOOLEAN_OBJ,
			Value:          fmt.Sprintf("(%s >= %s)", left.String(), right.String())}
	case operator == "<=":
		return &object.Expression{
			ExpressionType: object.BOOLEAN_OBJ,
			Value:          fmt.Sprintf("(%s >= %s)", left.String(), right.String())}
	case operator == "AND":
		return &object.Expression{
			ExpressionType: object.BOOLEAN_OBJ,
			Value:          fmt.Sprintf("(%s AND %s)", left.String(), right.String())}
	case operator == "OR":
		return &object.Expression{
			ExpressionType: object.BOOLEAN_OBJ,
			Value:          fmt.Sprintf("(%s OR %s)", left.String(), right.String())}
	default:
		return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
	}
}

// TODO: Check for Nullable
func evalInfixNullExpression(operator string, ref, null object.Object) object.Object {
	switch {
	case ref.Type() == object.NULL_OBJ:
		return newError("NULL cannot be compared to NULL")
	case operator == "==":
		return &object.Expression{
			ExpressionType: object.BOOLEAN_OBJ,
			Value:          fmt.Sprintf("(%s IS %s)", ref.String(), null.String())}
	case operator == "!=":
		return &object.Expression{
			ExpressionType: object.BOOLEAN_OBJ,
			Value:          fmt.Sprintf("(%s IS NOT %s)", ref.String(), null.String())}
	default:
		return newError("unknown operator: %s %s %s", ref.Type(), operator, null.Type())
	}
}

// Evaluate @ for expressions
func evalColumnCall(node ast.Node, ir *IR) object.Object {

	if node.TokenLiteral() != "@" {
		return newError("Invalid Token Literal. got=%s, want=%s", node.TokenLiteral(), "@")
	}

	if ir.currentField == nil {
		return newError("Invalid Column Call, %s", "No current field found")
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
