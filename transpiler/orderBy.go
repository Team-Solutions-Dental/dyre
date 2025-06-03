package transpiler

import (
	"github.com/vamuscari/dyre/ast"
	"github.com/vamuscari/dyre/object"
	"github.com/vamuscari/dyre/sql"
	"github.com/vamuscari/dyre/utils"
)

type orderByStatement struct {
	name string
}

// PrimaryIR can use OrderBy
// OrderBy will use a similar syntax to Querys where a column is fieldname: along with any provided expressions
// OrderBy should search current select fields and match the column name or alias

func (pir *PrimaryIR) OrderBy(string) {

}

func evalOrderBy(node ast.Node, ir *IR) object.Object {
	switch node := node.(type) {
	case *ast.RequestStatements:
		return evalOrderByStatements(node, ir)
	case *ast.ColumnStatement:
		return evalOrderByColumnStatement(node, ir)
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

func evalOrderByStatements(node *ast.RequestStatements, ir *IR) object.Object {
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

func evalOrderByColumnStatement(node *ast.ColumnStatement, ir *IR) object.Object {
	if !utils.Array_Contains(ir.FieldNames(), node.TokenLiteral()) {
		return newError("Requested Order By column %s not found for %s", node.TokenLiteral(), ir.endpoint.TableName)
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
