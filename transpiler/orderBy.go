package transpiler

import (
	"github.com/vamuscari/dyre/ast"
	"github.com/vamuscari/dyre/object"
	"github.com/vamuscari/dyre/sql"
)

// PrimaryIR can use OrderBy
// OrderBy will use a similar syntax to Querys where a column is fieldname: along with any provided expressions
// OrderBy should search current select fields and match the column name or alias

func (pir *PrimaryIR) OrderBy(req string) error {
	ast, err := parse(req)
	if err != nil {
		pir.error = err
		return err
	}

	pir.orderByAST = ast

	return nil
}

func evalOrderBy(node ast.Node, ir *IR) object.Object {
	switch node := node.(type) {
	case *ast.RequestStatements:
		return evalOrderByStatements(node, ir)
	case *ast.ColumnLiteral:
		return evalOrderByColumnLiteral(node, ir)
	case *ast.ExpressionStatement:
		return evalOrderByExpressionStatement(node, ir)
	case *ast.OrderExpression:
		return evalOrderByExpression(node, ir)
	default:
		return newError("Unknown Order By Evaluation Type: %T", node)
	}
}

// Eval Order By Request
func evalOrderByStatements(node *ast.RequestStatements, ir *IR) object.Object {
	var result object.Object

	for _, statement := range node.Statements {
		result = evalOrderBy(statement, ir)

		switch result := result.(type) {
		case *object.Error:
			return result
		}
	}
	return result
}

// Check for select statement and check if exists as field on endpoint
// Assign select statement accordingly
// Ex. ColumnName:
func evalOrderByColumnLiteral(node *ast.ColumnLiteral, ir *IR) object.Object {

	loc := ir.sql.SelectStatementLocation(node.TokenLiteral())
	field, ok := ir.endpoint.Fields[node.TokenLiteral()]

	if loc == -1 && !ok {
		return newError("Order By '%s' not found", node.TokenLiteral())
	}

	if loc >= 0 {
		ir.currentSelectStatement = ir.sql.SelectStatements[loc]
	} else if ok {
		ir.currentSelectStatement = field.SelectStatement()
	}

	// Add Order By.
	// If an expression is never called an Order By is still added
	ir.sql.OrderBy = append(
		ir.sql.OrderBy,
		&sql.OrderByStatement{Ascending: true, FieldName: ir.currentSelectStatement.Name()},
	)

	return nil
}

// Eval Expression statement.
// Ex. expression;
func evalOrderByExpressionStatement(stmnt *ast.ExpressionStatement, ir *IR) object.Object {

	evaluated := evalOrderBy(stmnt.Expression, ir)
	if isError(evaluated) {
		return evaluated
	}

	return nil

}

func evalOrderByExpression(node *ast.OrderExpression, ir *IR) object.Object {
	if len(ir.sql.OrderBy) < 1 {
		return nil
	}

	ir.sql.OrderBy[len(ir.sql.OrderBy)-1].Ascending = node.Ascending

	return nil
}
