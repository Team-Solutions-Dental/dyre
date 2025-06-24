package transpiler

import (
	"errors"
	"fmt"
	"strings"

	"github.com/vamuscari/dyre/ast"
	"github.com/vamuscari/dyre/endpoint"
	"github.com/vamuscari/dyre/lexer"
	"github.com/vamuscari/dyre/object"
	"github.com/vamuscari/dyre/object/objectRef"
	"github.com/vamuscari/dyre/object/objectType"
	"github.com/vamuscari/dyre/parser"
	"github.com/vamuscari/dyre/sql"
	"github.com/vamuscari/dyre/utils"
)

// Intermediate Representation
type IR struct {
	endpoint               *endpoint.Endpoint
	currentSelectStatement sql.SelectStatement
	ast                    *ast.RequestStatements
	sql                    *sql.Query
	isGroup                *bool
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
		sql:   &sql.Query{Depth: 1}}}
	return &ir, err
}

// Check if ir is group. If nil set value
func (ir *IR) checkGroup(expected bool) bool {
	if ir.isGroup == nil {
		ir.isGroup = &expected
		return true
	}

	if expected == *ir.isGroup {
		return true
	}

	return false
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

// Eval Table Query
// Evaluate only top level query since this runs recursivly.
// Evaluates joins then evalueates parent and adds fields into parent
func (ir *IR) evalTable() object.Object {
	if ir.endpoint.SchemaName != "" {
		ir.sql.From = ir.endpoint.SchemaName + "." + ir.endpoint.TableName
	} else {
		ir.sql.From = ir.endpoint.TableName
	}

	ir.sql.TableName = ir.endpoint.TableName

	// Eval Joins Before Parent
	for _, js := range ir.joins {
		result := js.childIR.evalTable()
		if isError(result) {
			return result
		}
	}

	result := eval(ir.ast, ir, nil)
	if isError(result) {
		return result
	}

	// Add statements from joins into parent.
	for _, j := range ir.joins {
		for _, ss := range j.childIR.sql.SelectStatements {
			// Ignore child joined on since parent should be referenced instead
			if ss.Name() == j.childOn {
				continue
			}
			fieldName := ss.Name()
			joinedSelect := sql.SelectField{
				FieldName: &fieldName,
				TableName: &j.alias,
				ObjType:   ss.ObjectType(),
			}
			ir.sql.SelectStatements = append(ir.sql.SelectStatements, &joinedSelect)
		}

		err := j.Check()
		if err != nil {
			return newError("%s", err.Error())
		}
	}

	return nil
}

func eval(node ast.Node, ir *IR, local *objectRef.LocalReferences) object.Object {
	switch node := node.(type) {
	case *ast.RequestStatements:
		return evalQueryStatements(node, ir)
	case *ast.ColumnLiteral:
		return evalColumnLiteral(node, ir)
	case *ast.ColumnFunction:
		return evalColumnFunction(node, ir)
	case *ast.GroupFunction:
		return evalGroupFunction(node, ir)
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
		right := eval(node.Right, ir, local)
		if isError(right) {
			return right
		}
		return evalPrefixExpression(node.Operator, right, local)
	case *ast.InfixExpression:
		left := eval(node.Left, ir, local)
		if isError(left) {
			return left
		}
		right := eval(node.Right, ir, local)
		if isError(right) {
			return right
		}
		return evalInfixExpression(node.Operator, left, right, local)
	case *ast.Reference:
		return evalColumnCall(node, ir, local)
	case *ast.CallExpression:
		return evalCallExpression(node.Function.TokenLiteral(), node.Arguments, ir, local)
	default:
		return newError("Unknown Evaluation Type: %T", node)
	}
}

func evalQueryStatements(node *ast.RequestStatements, ir *IR) object.Object {
	var result object.Object

	for _, statement := range node.Statements {
		result = eval(statement, ir, nil)
		switch result := result.(type) {
		case *object.Error:
			return result
		}
	}
	return result
}

func evalColumnLiteral(node *ast.ColumnLiteral, ir *IR) object.Object {
	if !ir.checkGroup(false) {
		return newError("Column '%s' cannot be called on Grouped Table '%s'", node.TokenLiteral(), ir.endpoint.TableName)
	}
	if !utils.Array_Contains(ir.endpoint.FieldNames, node.TokenLiteral()) {
		return newError("Requested column %s not found for %s", node.TokenLiteral(), ir.endpoint.TableName)
	}

	column := ir.endpoint.Fields[node.TokenLiteral()]

	selectStatementLoc := ir.sql.SelectStatementLocation(column.Name)
	if selectStatementLoc >= 0 {
		newSelectStatementList := []sql.SelectStatement{}
		for i, ss := range ir.sql.SelectStatements {
			if i != selectStatementLoc {
				newSelectStatementList = append(newSelectStatementList, ss)
			}
		}

		// Append duplicate column to end of list
		newSelectStatementList = append(newSelectStatementList, ir.sql.SelectStatements[selectStatementLoc])
		ir.currentSelectStatement = ir.sql.SelectStatements[selectStatementLoc]
		ir.sql.SelectStatements = newSelectStatementList

	} else {
		selects := &sql.SelectField{
			FieldName: &column.Name,
			TableName: &ir.endpoint.TableName,
			ObjType:   column.FieldType,
		}
		ir.currentSelectStatement = selects
		ir.sql.SelectStatements = append(ir.sql.SelectStatements, selects)
	}

	return nil
}

func evalColumnFunction(node *ast.ColumnFunction, ir *IR) object.Object {
	if !ir.checkGroup(false) {
		return newError("Column '%s' cannot be called on Grouped Table '%s'", node.Fn, ir.endpoint.TableName)
	}
	args := evalExpressions(node.Arguments, ir)

	_, ok := columnFunctions[node.Fn]
	if !ok {
		return newError("Column Function '%s' not found", node.Fn)
	}

	return columnFunctions[node.Fn](ir, args...)
}

func evalGroupFunction(node *ast.GroupFunction, ir *IR) object.Object {
	if !ir.checkGroup(true) {
		return newError("Group '%s' cannot be called on Non-Grouped Table '%s'", node.Fn, ir.endpoint.TableName)
	}
	args := evalExpressions(node.Arguments, ir)

	_, ok := groupFunctions[node.Fn]
	if !ok {
		return newError("Group Function Function '%s' not found", node.Fn)
	}

	return groupFunctions[node.Fn](ir, args...)
}

func evalExpressionStatement(
	stmnt *ast.ExpressionStatement,
	ir *IR,
) object.Object {

	if stmnt.Expression == nil {
		return nil
	}

	local := objectRef.NewLocalReferences()

	evaluated := eval(stmnt.Expression, ir, local)
	if isError(evaluated) {
		return evaluated
	}

	highest := local.Highest()

	if highest == -1 {
		return nil
	}

	if !local.AllSame() {
		return newError("Not all references are the same type")
	}

	switch highest {
	case objectRef.FIELD:
		if evaluated.Type() == objectType.BOOLEAN {
			ir.sql.WhereStatements = append(ir.sql.WhereStatements, evaluated.String())
		}
	case objectRef.EXPRESSION:
		if evaluated.Type() == objectType.BOOLEAN {
			ir.sql.AliasWhereStatements = append(ir.sql.AliasWhereStatements, evaluated.String())
		}
	case objectRef.GROUP:
		if evaluated.Type() == objectType.BOOLEAN {
			ir.sql.HavingStatements = append(ir.sql.HavingStatements, evaluated.String())
		}
	}

	return nil

}

func evalPrefixExpression(
	operator string,
	right object.Object,
	local *objectRef.LocalReferences,
) object.Object {
	switch operator {
	case "!":
		return evalBangOperatorExpression(right, local)
	case "-":
		return evalMinusPrefixOperatorExpression(right, local)
	default:
		return newError("unknown operator: %s%s", operator, right.Type())
	}
}

func evalBangOperatorExpression(
	right object.Object,
	local *objectRef.LocalReferences,
) object.Object {
	switch {
	case right.Type() == objectType.BOOLEAN:
		return &object.Expression{
			ExpressionType: objectType.BOOLEAN,
			Value:          fmt.Sprintf("!%s", right.String())}
	default:
		return newError("Invalid Bang Operator Expression %s", right.String())
	}
}

func evalMinusPrefixOperatorExpression(
	right object.Object,
	local *objectRef.LocalReferences,
) object.Object {
	switch {
	case right.Type() == objectType.INTEGER:
		return &object.Expression{
			ExpressionType: objectType.INTEGER,
			Value:          fmt.Sprintf("-%s", right.String())}
	default:
		return newError("Invalid Minus Prefix Operator Expression %s", right.String())
	}
}

func evalInfixExpression(
	operator string,
	left, right object.Object,
	local *objectRef.LocalReferences,
) object.Object {
	switch {
	case left.Type() == objectType.NULL:
		return evalInfixNullExpression(operator, right, left, local)
	case right.Type() == objectType.NULL:
		return evalInfixNullExpression(operator, left, right, local)
	// case left.Type() == objectType.NULL || right.Type() == object.NULL_OBJ:
	// 	return newError("type mismatch: %s %s %s", left.Type(), operator, right.Type())
	case operator == "==":
		return &object.Expression{
			ExpressionType: objectType.BOOLEAN,
			Value:          fmt.Sprintf("(%s = %s)", left.String(), right.String())}
	case operator == "!=":
		return &object.Expression{
			ExpressionType: objectType.BOOLEAN,
			Value:          fmt.Sprintf("(%s != %s)", left.String(), right.String())}
	case operator == ">":
		return &object.Expression{
			ExpressionType: objectType.BOOLEAN,
			Value:          fmt.Sprintf("(%s > %s)", left.String(), right.String())}
	case operator == "<":
		return &object.Expression{
			ExpressionType: objectType.BOOLEAN,
			Value:          fmt.Sprintf("(%s < %s)", left.String(), right.String())}
	case operator == ">=":
		return &object.Expression{
			ExpressionType: objectType.BOOLEAN,
			Value:          fmt.Sprintf("(%s >= %s)", left.String(), right.String())}
	case operator == "<=":
		return &object.Expression{
			ExpressionType: objectType.BOOLEAN,
			Value:          fmt.Sprintf("(%s >= %s)", left.String(), right.String())}
	case operator == "AND":
		return &object.Expression{
			ExpressionType: objectType.BOOLEAN,
			Value:          fmt.Sprintf("(%s AND %s)", left.String(), right.String())}
	case operator == "OR":
		return &object.Expression{
			ExpressionType: objectType.BOOLEAN,
			Value:          fmt.Sprintf("(%s OR %s)", left.String(), right.String())}
	default:
		return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
	}
}

// TODO: Check for Nullable
func evalInfixNullExpression(
	operator string,
	ref, null object.Object,
	local *objectRef.LocalReferences,
) object.Object {
	switch {
	case ref.Type() == objectType.NULL:
		return newError("NULL cannot be compared to NULL")
	case operator == "==":
		return &object.Expression{
			ExpressionType: objectType.BOOLEAN,
			Value:          fmt.Sprintf("(%s IS %s)", ref.String(), null.String())}
	case operator == "!=":
		return &object.Expression{
			ExpressionType: objectType.BOOLEAN,
			Value:          fmt.Sprintf("(%s IS NOT %s)", ref.String(), null.String())}
	default:
		return newError("unknown operator: %s %s %s", ref.Type(), operator, null.Type())
	}
}

// Evaluate @ for expressions
// WARN: Cannot call alias
func evalColumnCall(
	node *ast.Reference,
	ir *IR,
	local *objectRef.LocalReferences,
) object.Object {

	if node.Argument == nil && ir.currentSelectStatement == nil {
		return newError("Invalid Column Call, %s", "No current field specified or referenced")
	}

	if local == nil {
		return newError("Invalid Column Call '%s', %s", node.String(), "Missing Local References")
	}

	if node.Argument == nil {
		switch ir.currentSelectStatement.Type() {
		case "FIELD":
			local.Set(ir.currentSelectStatement.Name(), objectRef.FIELD)
			return &object.Expression{ExpressionType: ir.currentSelectStatement.ObjectType(),
				Value: fmt.Sprintf("%s.[%s]", ir.endpoint.TableName, ir.currentSelectStatement.Name())}
		case "EXPRESSION":
			local.Set(ir.currentSelectStatement.Name(), objectRef.EXPRESSION)
			se := ir.currentSelectStatement.(*sql.SelectExpression)
			return se.Expression
		case "GROUP_FIELD":
			local.Set(ir.currentSelectStatement.Name(), objectRef.GROUP)
			return &object.Expression{ExpressionType: ir.currentSelectStatement.ObjectType(),
				Value: fmt.Sprintf("%s.[%s]", ir.endpoint.TableName, ir.currentSelectStatement.Name())}
		case "GROUP_EXPRESSION":
			local.Set(ir.currentSelectStatement.Name(), objectRef.GROUP)
			ge := ir.currentSelectStatement.(*sql.SelectGroupExpression)
			return ge.Expression
		}
	}

	eval := eval(node.Argument, ir, local)

	if isError(eval) {
		return eval
	}

	if eval.Type() != objectType.STRING {
		return newError("Invalid Column Call Expression, type not string. got=%s", eval.Type())
	}

	str := eval.(*object.String)

	field, ok := ir.endpoint.Fields[str.Value]
	if !ok {
		return newError("Invalid Column Call Expression, type not string. got=%s", eval.Type())
	}

	local.Set(field.Name, objectRef.FIELD)
	return &object.Expression{ExpressionType: field.FieldType,
		Value: fmt.Sprintf("%s.[%s]", ir.endpoint.TableName, str.Value)}

}

func newError(format string, a ...interface{}) *object.Error {
	return &object.Error{Message: fmt.Sprintf(format, a...)}
}

func isError(obj object.Object) bool {
	if obj != nil {
		return obj.Type() == objectType.ERROR
	}
	return false
}

func evalExpressions(
	exps []ast.Expression,
	ir *IR,
) []object.Object {
	var results []object.Object

	for _, e := range exps {
		local := objectRef.NewLocalReferences()
		evaluated := eval(e, ir, local)
		results = append(results, evaluated)
	}

	return results
}

func evalCallExpression(
	function string,
	exps []ast.Expression,
	ir *IR,
	local *objectRef.LocalReferences,
) object.Object {
	args := evalExpressions(exps, ir)

	_, ok := builtins[function]
	if !ok {
		return newError("Function %s not found", function)
	}

	return builtins[function](ir, args...)
}
