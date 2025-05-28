package transpiler

import (
	"errors"
	"fmt"
	"strings"

	"github.com/vamuscari/dyre/ast"
	"github.com/vamuscari/dyre/endpoint"
	"github.com/vamuscari/dyre/lexer"
	"github.com/vamuscari/dyre/parser"
	"github.com/vamuscari/dyre/utils"
)

func New(query string, endpoint *endpoint.Endpoint) *IR {
	l := lexer.New(query)
	p := parser.New(l)
	q := p.ParseQuery()
	var ir IR = IR{endpoint: endpoint, ast: q}
	ir.eval()
	return &ir
}

// Intermediate Representation
type IR struct {
	endpoint         *endpoint.Endpoint
	ast              *ast.QueryStatements
	currentField     *endpoint.Field
	fields           []*endpoint.Field
	selectStatements []*selectStatement
	limit            *int
	whereStatements  []string
	joinStatements   []*joinStatement
	Errors           []string
}

// TODO: reorder columns for select if found
func (ir *IR) eval() {

	for _, statement := range ir.ast.Columns {
		if !utils.Array_Contains(ir.endpoint.FieldNames, statement.TokenLiteral()) {
			ir.Errors = append(ir.Errors, "column not found: "+statement.TokenLiteral())
		}

		column := ir.endpoint.Fields[statement.TokenLiteral()]
		ir.currentField = &column
		ir.fields = append(ir.fields, ir.currentField)

		ir.evalColumnStatement(statement)

		selectStatementLoc := ir.selectStatementLocation(ir.currentField.Name)
		if selectStatementLoc >= 0 {
			newSelectStatement := []*selectStatement{}
			for i, ss := range ir.selectStatements {
				if i != selectStatementLoc {
					newSelectStatement = append(newSelectStatement, ss)
				}
			}

			newSelectStatement = append(newSelectStatement, ir.selectStatements[selectStatementLoc])
			ir.selectStatements = newSelectStatement

		} else {
			selects := &selectStatement{fieldName: &ir.currentField.Name, tableName: &ir.endpoint.TableName}
			ir.selectStatements = append(ir.selectStatements, selects)
		}
	}
}

func (ir *IR) evalColumnStatement(node ast.Node) {
	statement, ok := node.(*ast.ColumnStatement)
	if !ok {
		ir.Errors = append(ir.Errors, "Invalid statement")
		return
	}

	exp := statement.Expression
	if exp == nil {
		return
	}

	eval, err := ir.evalExpression(exp)
	if err != nil {
		ir.Errors = append(ir.Errors, err.Error())
	}

	ir.whereStatements = append(ir.whereStatements, eval)

	return
}

func (ir *IR) evalExpression(node ast.Node) (string, error) {
	switch node := node.(type) {
	case *ast.IntegerLiteral:
		return node.TokenLiteral(), nil
	case *ast.NullLiteral:
		return "NULL", nil
	case *ast.StringLiteral:
		return fmt.Sprintf("'%s'", node.TokenLiteral()), nil
	case *ast.Boolean:
		return nativeBooltoSQLBool(node.Value), nil
	case *ast.Identifier:
		return "Identifier", nil
	case *ast.PrefixExpression:
		right, err := ir.evalExpression(node.Right)
		if err != nil {
			return "", err
		}
		return evalPrefixExpression(node.Operator, right)
	case *ast.InfixExpression:
		left, err := ir.evalExpression(node.Left)
		if err != nil {
			return "", err
		}
		right, err := ir.evalExpression(node.Right)
		if err != nil {
			return "", err
		}
		return evalInfixExpression(node.Operator, left, right), nil
	case *ast.ColumnCall:
		return ir.evalColumnCall(node)
	case *ast.CallExpression:
		return ir.evalCallExpression(node.Function.TokenLiteral(), node.Arguments)
	default:
		return "", errors.New("Unknown Expression")
	}
}

func nativeBooltoSQLBool(input bool) string {
	if input == true {
		return "1"
	}
	return "0"
}

func evalPrefixExpression(operator, right string) (string, error) {
	switch operator {
	case "!":
		return fmt.Sprintf("!%s", right), nil
	case "-":
		return fmt.Sprintf("-%s", right), nil
	default:
		return "", errors.New(fmt.Sprintf("unknown operator: %s", operator))
	}
}

func evalInfixExpression(operator, left, right string) string {
	switch {
	case operator == "==":
		return fmt.Sprintf("(%s = %s)", left, right)
	case operator == "!=":
		return fmt.Sprintf("(%s != %s)", left, right)
	case operator == ">":
		return fmt.Sprintf("(%s > %s)", left, right)
	case operator == "<":
		return fmt.Sprintf("(%s < %s)", left, right)
	case operator == ">=":
		return fmt.Sprintf("(%s >= %s)", left, right)
	case operator == "<=":
		return fmt.Sprintf("(%s <= %s)", left, right)
	case operator == "&&":
		return fmt.Sprintf("(%s AND %s)", left, right)
	case operator == "||":
		return fmt.Sprintf("(%s OR %s)", left, right)
	default:
		return ""
	}
}

func (ir *IR) evalColumnCall(node ast.Node) (string, error) {

	if node.TokenLiteral() != "@" {
		return "", errors.New(fmt.Sprintf("Invalid Token Literal. got=%s, want=%s", node.TokenLiteral(), "@"))
	}

	return (ir.endpoint.TableName + "." + ir.currentField.Name), nil
}

func (ir *IR) evalExpressions(
	exps []ast.Expression,
) ([]string, error) {
	var result []string

	for _, e := range exps {
		evaluated, err := ir.evalExpression(e)
		if err != nil {
			return nil, err
		}
		result = append(result, evaluated)
	}

	return result, nil
}

func (ir *IR) evalCallExpression(function string, exps []ast.Expression) (string, error) {
	args, err := ir.evalExpressions(exps)
	if err != nil {
		return "", err
	}

	return builtins[function](ir, args...)
}

func (ir *IR) BuildSQLQuery() string {
	var query string = "SELECT "

	if ir.limit != nil && *ir.limit > 0 {
		query = query + fmt.Sprintf("TOP %d ", *ir.limit)
	}

	query = query + selectConstructor(ir.selectStatements)

	if ir.endpoint.SchemaName != "" {
		query = query + " FROM " + ir.endpoint.SchemaName + "." + ir.endpoint.TableName
	} else {
		query = query + " FROM " + ir.endpoint.TableName
	}

	if len(ir.joinStatements) > 0 {
		query = query + joinConstructor(ir.joinStatements)
	}

	if len(ir.whereStatements) > 0 {
		query = query + whereConstructor(ir.whereStatements)
	}

	return query
}

func selectConstructor(selects []*selectStatement) string {
	var selectStrings []string
	for _, ss := range selects {
		selectStrings = append(selectStrings, ss.String())
	}

	return strings.Join(selectStrings, ", ")
}

func joinConstructor(joins []*joinStatement) string {
	var joinArr []string
	for _, j := range joins {
		if j.errors != nil {
			fmt.Println("skip")
			continue
		}

		joinArr = append(joinArr, fmt.Sprintf(" %s JOIN ( %s ) AS %s ON %s = %s", j.Type, j.ir.BuildSQLQuery(), j.endpoint.TableName, j.parentIrOn(), j.joinIrOn()))
	}

	return strings.Join(joinArr, " ")
}

func whereConstructor(statements []string) string {
	where := ""
	if len(statements) < 1 {
		return where
	}
	if len(statements) == 1 {
		where = fmt.Sprintf(" WHERE %s", statements[0])
		return where
	}
	where = fmt.Sprintf(" WHERE %s", statements[0])
	for i := 1; i < len(statements); i++ {
		where = where + " AND " + statements[i]
	}
	return where
}

func (ir *IR) INNERJOIN(req string) *joinType {
	join := &joinType{Type: "INNER", parentIR: ir}

	endpoint, err := ir.endpoint.Service.GetEndpoint(req)
	if err != nil {
		join.errors = append(join.errors, err)
	}

	join.endpoint = endpoint

	return join

}

func (ir *IR) LEFTJOIN(req string) *joinType {
	join := &joinType{Type: "LEFT", parentIR: ir}

	endpoint, err := ir.endpoint.Service.GetEndpoint(req)
	if err != nil {
		join.errors = append(join.errors, err)
	}

	join.endpoint = endpoint

	return join
}

func (ir *IR) FieldNames() []string {

	var names []string
	for _, f := range ir.fields {
		names = append(names, f.Name)
	}

	return names
}

func (ir *IR) selectStatementLocation(input string) int {
	for i, ss := range ir.selectStatements {
		if *ss.fieldName == input {
			return i
		}
	}
	return -1
}

func (ir *IR) LIMIT(input int) *IR {
	ir.limit = &input
	return ir
}
