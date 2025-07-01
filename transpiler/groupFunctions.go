package transpiler

import (
	"github.com/vamuscari/dyre/object"
	"github.com/vamuscari/dyre/object/objectRef"
	"github.com/vamuscari/dyre/object/objectType"
	"github.com/vamuscari/dyre/sql"
)

var groupFunctions = map[string]func(ir *IR, local *objectRef.LocalReferences, args ...object.Object) object.Object{
	// GROUP(ColumnName:string)
	// GROUP(Alias:string, Expression)
	"GROUP": func(ir *IR, local *objectRef.LocalReferences, args ...object.Object) object.Object {
		if len(args) == 1 {
			return groupColumn(ir, local, args...)
		}

		if len(args) == 2 {
			return groupExpression(ir, local, args...)
		}

		return newError("wrong number of arguments. got=%d, want=1-2", len(args))

	},
	// COUNT(alias: string, input: expression)
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
			Value:          expression.String(),
		}

		expr := &sql.SelectGroupExpression{Query: ir.sql, Fn: &fn, Alias: &name_obj.Value, Expression: out}

		local.Set(expr.Statement(), objectRef.GROUP)

		ir.currentSelectStatement = expr
		ir.sql.SelectStatements = append(ir.sql.SelectStatements, expr)

		return nil
	},
	// SUM(name: string, input: expression)
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
			Value:          expression.String(),
		}

		expr := &sql.SelectGroupExpression{Query: ir.sql, Fn: &fn, Alias: &name_obj.Value, Expression: out}
		local.Set(expr.Statement(), objectRef.GROUP)

		ir.currentSelectStatement = expr
		ir.sql.SelectStatements = append(ir.sql.SelectStatements, expr)

		return nil
	},
	//AVG(name: string, input: expression)
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
			Value:          expression.String(),
		}

		expr := &sql.SelectGroupExpression{Query: ir.sql, Fn: &fn, Alias: &name_obj.Value, Expression: out}
		local.Set(expr.Statement(), objectRef.GROUP)

		ir.currentSelectStatement = expr
		ir.sql.SelectStatements = append(ir.sql.SelectStatements, expr)

		return nil
	},
	// MIN(alias: string, input: expression)
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
			Value:          expression.String(),
		}

		expr := &sql.SelectGroupExpression{Query: ir.sql, Fn: &fn, Alias: &name_obj.Value, Expression: out}
		local.Set(expr.Statement(), objectRef.GROUP)

		ir.currentSelectStatement = expr
		ir.sql.SelectStatements = append(ir.sql.SelectStatements, expr)

		return nil
	},
	// MAX(alias: string, input: expression)
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
			Value:          expression.String(),
		}

		expr := &sql.SelectGroupExpression{Query: ir.sql, Fn: &fn, Alias: &name_obj.Value, Expression: out}
		local.Set(expr.Statement(), objectRef.GROUP)

		ir.currentSelectStatement = expr
		ir.sql.SelectStatements = append(ir.sql.SelectStatements, expr)

		return nil
	},
}

func groupColumn(ir *IR, local *objectRef.LocalReferences, args ...object.Object) object.Object {

	name, oerr := object.CastType[*object.String](args[0])
	if oerr != nil {
		return oerr
	}

	selectStatementLoc := ir.sql.SelectStatementLocation(name.Value)
	if selectStatementLoc >= 0 {
		return newError("Cannot group already defined field '%s'", name.Value)
	}

	//groupSelect := &sql.SelectGroupField{FieldName: &name.Value, TableName: &ir.endpoint.TableName}
	groupSelect := &sql.SelectGroupField{Query: ir.sql}

	field, field_ok := ir.endpoint.Fields[name.Value]
	joined, joined_ok := ir.sql.GetJoinedStatement(name.Value)
	if field_ok {
		local.Set(field.Name, objectRef.GROUP)
		groupSelect.Query = ir.sql
		groupSelect.FieldName = &name.Value
		groupSelect.TableName = &ir.endpoint.Name
		groupSelect.ObjType = field.Type()
		groupSelect.HasNull = field.Nullable
	} else if joined_ok {
		local.Set(joined.Statement(), objectRef.GROUP)
		groupSelect.Query = ir.sql
		groupSelect.FieldName = joined.FieldName
		groupSelect.TableName = joined.TableName
		groupSelect.ObjType = joined.ObjType
		groupSelect.HasNull = joined.Nullable()
	} else {
		return newError("Column '%s' not found", name.Value)
	}

	ir.currentSelectStatement = groupSelect
	ir.sql.SelectStatements = append(ir.sql.SelectStatements, groupSelect)
	ir.sql.GroupByStatements = append(ir.sql.GroupByStatements, groupSelect.Statement())

	return nil
}

func groupExpression(ir *IR, local *objectRef.LocalReferences, args ...object.Object) object.Object {
	alias, oerr := object.CastType[*object.String](args[0])
	if oerr != nil {
		return oerr
	}

	fn := ""
	groupSelect := &sql.SelectGroupExpression{
		Query:      ir.sql,
		Fn:         &fn,
		Alias:      &alias.Value,
		Expression: args[1],
		HasNull:    args[1].Nullable(),
	}
	local.Set(args[1].String(), objectRef.GROUP)
	ir.currentSelectStatement = groupSelect
	ir.sql.SelectStatements = append(ir.sql.SelectStatements, groupSelect)
	ir.sql.GroupByStatements = append(ir.sql.GroupByStatements, args[1].String())

	return nil
}
