package transpiler

import (
	"github.com/vamuscari/dyre/object"
	"github.com/vamuscari/dyre/object/objectType"
	"github.com/vamuscari/dyre/sql"
)

var columnFunctions = map[string]func(ir *IR, args ...object.Object) object.Object{
	// AS(name, expression)
	"AS": func(ir *IR, args ...object.Object) object.Object {
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

		expr := &sql.SelectExpression{
			Query:      ir.sql,
			Alias:      &name_obj.Value,
			Expression: expression,
			HasNull:    expression.Nullable(),
		}

		ir.currentSelectStatement = expr
		ir.sql.SelectStatements = append(ir.sql.SelectStatements, expr)

		return nil
	},
	// EXCLUDE(name)
	// Create Current Select Statement but do not include it in sql representation
	"EXCLUDE": func(ir *IR, args ...object.Object) object.Object {
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
			return newError("Cannot exclude already defined field '%s'", string_obj.Value)
		}

		ir.currentSelectStatement = &sql.SelectField{
			Query:     ir.sql,
			FieldName: &string_obj.Value,
			TableName: &ir.endpoint.TableName,
		}
		return nil
	},
}
