package transpiler

import (
	"fmt"

	"github.com/vamuscari/dyre/object"
	"github.com/vamuscari/dyre/object/objectRef"
	"github.com/vamuscari/dyre/object/objectType"
	"github.com/vamuscari/dyre/sql"
)

var groupFunctions = map[string]func(ir *IR, local *objectRef.LocalReferences, args ...object.Object) object.Object{
	// GROUP(name)
	"GROUP": func(ir *IR, local *objectRef.LocalReferences, args ...object.Object) object.Object {
		if len(args) != 1 {
			return newError("wrong number of arguments. got=%d, want=1", len(args))
		}

		name := args[0]

		if name.Type() != objectType.STRING {
			return newError("Invalid name identity type, got=%s, want=STRING", name.Type())
		}

		string_obj, ok := name.(*object.String)
		if !ok {
			return newError("Invalid name identity type convertion, got=%s, want=STRING", name.Type())
		}

		selectStatementLoc := ir.sql.SelectStatementLocation(string_obj.Value)
		if selectStatementLoc >= 0 {
			return newError("Cannot group already defined field '%s'", string_obj.Value)
		}

		groupSelect := &sql.SelectGroupField{FieldName: &string_obj.Value, TableName: &ir.endpoint.TableName}
		local.Set(groupSelect.Statement(), objectRef.GROUP)
		ir.currentSelectStatement = groupSelect
		ir.sql.SelectStatements = append(ir.sql.SelectStatements, groupSelect)

		ir.sql.GroupByStatements = append(ir.sql.GroupByStatements, groupSelect.Statement())

		return nil
	},
	// COUNT(name, expression)
	"COUNT": func(ir *IR, local *objectRef.LocalReferences, args ...object.Object) object.Object {
		fn := "COUNT"
		if len(args) != 2 {
			return newError("wrong number of arguments. got=%d, want=2", len(args))
		}

		name := args[0]
		expression := args[1]

		if name.Type() != objectType.STRING {
			return newError("Invalid name identity type, got=%s, want=STRING", name.Type())
		}

		name_obj, ok := name.(*object.String)
		if !ok {
			return newError("Invalid name identity type convertion, got=%s, want=STRING", name.Type())
		}

		selectStatementLoc := ir.sql.SelectStatementLocation(name_obj.Value)
		if selectStatementLoc >= 0 {
			return newError("Field '%s' name already defined ", name_obj.Value)
		}

		if isError(expression) {
			return expression
		}

		out := &object.Expression{
			ExpressionType: objectType.INTEGER,
			Value:          fmt.Sprintf("COUNT( %s )", expression.String()),
		}

		expr := &sql.SelectGroupExpression{Fn: &fn, Alias: &name_obj.Value, Expression: out}

		local.Set(expr.Statement(), objectRef.GROUP)

		ir.currentSelectStatement = expr
		ir.sql.SelectStatements = append(ir.sql.SelectStatements, expr)

		return nil
	},
	// SUM(name, expression)
	"SUM": func(ir *IR, local *objectRef.LocalReferences, args ...object.Object) object.Object {
		fn := "SUM"
		if len(args) != 2 {
			return newError("wrong number of arguments. got=%d, want=2", len(args))
		}

		name := args[0]
		expression := args[1]

		if name.Type() != objectType.STRING {
			return newError("Invalid name identity type, got=%s, want=STRING", name.Type())
		}

		name_obj, ok := name.(*object.String)
		if !ok {
			return newError("Invalid name identity type convertion, got=%s, want=STRING", name.Type())
		}

		selectStatementLoc := ir.sql.SelectStatementLocation(name_obj.Value)
		if selectStatementLoc >= 0 {
			return newError("Field '%s' name already defined ", name_obj.Value)
		}

		if isError(expression) {
			return expression
		}

		out := &object.Expression{
			ExpressionType: objectType.INTEGER,
			Value:          fmt.Sprintf("SUM( %s )", expression.String()),
		}

		expr := &sql.SelectGroupExpression{Fn: &fn, Alias: &name_obj.Value, Expression: out}
		local.Set(expr.Statement(), objectRef.GROUP)

		ir.currentSelectStatement = expr
		ir.sql.SelectStatements = append(ir.sql.SelectStatements, expr)

		return nil
	},
	"AVG": func(ir *IR, local *objectRef.LocalReferences, args ...object.Object) object.Object {
		fn := "AVG"
		if len(args) != 2 {
			return newError("wrong number of arguments. got=%d, want=2", len(args))
		}

		name := args[0]
		expression := args[1]

		if name.Type() != objectType.STRING {
			return newError("Invalid name identity type, got=%s, want=STRING", name.Type())
		}

		name_obj, ok := name.(*object.String)
		if !ok {
			return newError("Invalid name identity type convertion, got=%s, want=STRING", name.Type())
		}

		selectStatementLoc := ir.sql.SelectStatementLocation(name_obj.Value)
		if selectStatementLoc >= 0 {
			return newError("Field '%s' name already defined ", name_obj.Value)
		}

		if isError(expression) {
			return expression
		}

		out := &object.Expression{
			ExpressionType: objectType.INTEGER,
			Value:          fmt.Sprintf("AVG( %s )", expression.String()),
		}

		expr := &sql.SelectGroupExpression{Fn: &fn, Alias: &name_obj.Value, Expression: out}
		local.Set(expr.Statement(), objectRef.GROUP)

		ir.currentSelectStatement = expr
		ir.sql.SelectStatements = append(ir.sql.SelectStatements, expr)

		return nil
	},
	"MIN": func(ir *IR, local *objectRef.LocalReferences, args ...object.Object) object.Object {
		fn := "MIN"
		if len(args) != 2 {
			return newError("wrong number of arguments. got=%d, want=2", len(args))
		}

		name := args[0]
		expression := args[1]

		if name.Type() != objectType.STRING {
			return newError("Invalid name identity type, got=%s, want=STRING", name.Type())
		}

		name_obj, ok := name.(*object.String)
		if !ok {
			return newError("Invalid name identity type convertion, got=%s, want=STRING", name.Type())
		}

		selectStatementLoc := ir.sql.SelectStatementLocation(name_obj.Value)
		if selectStatementLoc >= 0 {
			return newError("Field '%s' name already defined ", name_obj.Value)
		}

		if isError(expression) {
			return expression
		}

		out := &object.Expression{
			ExpressionType: objectType.INTEGER,
			Value:          fmt.Sprintf("MIN( %s )", expression.String()),
		}

		expr := &sql.SelectGroupExpression{Fn: &fn, Alias: &name_obj.Value, Expression: out}
		local.Set(expr.Statement(), objectRef.GROUP)

		ir.currentSelectStatement = expr
		ir.sql.SelectStatements = append(ir.sql.SelectStatements, expr)

		return nil
	},
	"MAX": func(ir *IR, local *objectRef.LocalReferences, args ...object.Object) object.Object {
		fn := "MAX"
		if len(args) != 2 {
			return newError("wrong number of arguments. got=%d, want=2", len(args))
		}

		name := args[0]
		expression := args[1]

		if name.Type() != objectType.STRING {
			return newError("Invalid name identity type, got=%s, want=STRING", name.Type())
		}

		name_obj, ok := name.(*object.String)
		if !ok {
			return newError("Invalid name identity type convertion, got=%s, want=STRING", name.Type())
		}

		selectStatementLoc := ir.sql.SelectStatementLocation(name_obj.Value)
		if selectStatementLoc >= 0 {
			return newError("Field '%s' name already defined ", name_obj.Value)
		}

		if isError(expression) {
			return expression
		}

		out := &object.Expression{
			ExpressionType: objectType.INTEGER,
			Value:          fmt.Sprintf("MAX( %s )", expression.String()),
		}

		expr := &sql.SelectGroupExpression{Fn: &fn, Alias: &name_obj.Value, Expression: out}
		local.Set(expr.Statement(), objectRef.GROUP)

		ir.currentSelectStatement = expr
		ir.sql.SelectStatements = append(ir.sql.SelectStatements, expr)

		return nil
	},
}
