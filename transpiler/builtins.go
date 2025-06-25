package transpiler

import (
	"fmt"

	"github.com/vamuscari/dyre/object"
	"github.com/vamuscari/dyre/object/objectRef"
	"github.com/vamuscari/dyre/object/objectType"
)

var builtins = map[string]func(ir *IR, local *objectRef.LocalReferences, args ...object.Object) object.Object{
	"len": func(ir *IR, local *objectRef.LocalReferences, args ...object.Object) object.Object {
		if len(args) != 1 {
			return newError("wrong number of arguments. got=%d, want=1", len(args))
		}

		arg := args[0]

		switch {
		case arg.Type() == objectType.STRING:
			return &object.Expression{ExpressionType: objectType.INTEGER,
				Value: fmt.Sprintf("LEN(%s)", args[0])}
		default:
			return newError("Invalid Type. %s %s", arg.Type(), arg.String())
		}

	},
	"datepart": func(ir *IR, local *objectRef.LocalReferences, args ...object.Object) object.Object {
		if len(args) != 2 {
			return newError("wrong number of arguments. got=%d, want=2", len(args))
		}

		return &object.Expression{ExpressionType: objectType.DATE,
			Value: fmt.Sprintf("datepart(%s, %s)", args[0], args[1])}
	},
	"convert": func(ir *IR, local *objectRef.LocalReferences, args ...object.Object) object.Object {
		if len(args) < 2 {
			return newError("wrong number of arguments. got=%d, want=2-3", len(args))
		}

		if len(args) > 3 {
			return newError("wrong number of arguments. got=%d, want=2-3", len(args))
		}

		if len(args) == 2 {
			return &object.Expression{ExpressionType: objectType.DATE,
				Value: fmt.Sprintf("CONVERT(%s, %s)", args[0], args[1])}
		}

		if len(args) == 3 {
			return &object.Expression{ExpressionType: objectType.DATE,
				Value: fmt.Sprintf("CONVERT(%s, %s, %s)", args[0], args[1], args[2])}
		}

		return nil
	},
	"date": func(ir *IR, local *objectRef.LocalReferences, args ...object.Object) object.Object {
		if len(args) != 1 {
			return newError("wrong number of arguments. got=%d, want=1", len(args))
		}

		arg := args[0]

		switch {
		case arg.Type() == objectType.STRING:
			return &object.Expression{ExpressionType: objectType.DATE,
				Value: fmt.Sprintf("CONVERT(date, %s, 23)", args[0])}
		default:
			return newError("Invalid Type. %s %s", arg.Type(), arg.String())
		}
	},
	"datetime": func(ir *IR, local *objectRef.LocalReferences, args ...object.Object) object.Object {
		if len(args) != 1 {
			return newError("wrong number of arguments. got=%d, want=1", len(args))
		}

		arg := args[0]

		switch {
		case arg.Type() == objectType.STRING:
			return &object.Expression{ExpressionType: objectType.DATETIME,
				Value: fmt.Sprintf("CONVERT(date, %s, 127)", args[0])}
		default:
			return newError("Invalid Type. %s %s", arg.Type(), arg.String())
		}
	},
	//
	"like": func(ir *IR, local *objectRef.LocalReferences, args ...object.Object) object.Object {
		if len(args) != 2 {
			return newError("wrong number of arguments. got=%d, want=2", len(args))
		}

		column := args[0]
		comparison := args[1]

		switch {
		case column.Type() == objectType.STRING:
			return &object.Expression{ExpressionType: objectType.BOOLEAN,
				Value: fmt.Sprintf("(%s LIKE %s)", column.String(), comparison.String())}
		default:
			return newError("Invalid Type. %s %s", column.Type(), column.String())
		}
	},
}
