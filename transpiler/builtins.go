package transpiler

import (
	"fmt"

	"github.com/vamuscari/dyre/object"
)

var builtins = map[string]func(ir *IR, args ...object.Object) object.Object{
	"len": func(ir *IR, args ...object.Object) object.Object {
		if len(args) != 1 {
			return newError(fmt.Sprintf("wrong number of arguments. got=%d, want=1", len(args)))
		}

		arg := args[0]

		switch {
		case arg.Type() == object.STRING_OBJ:
			return &object.Expression{ExpressionType: object.INTEGER_OBJ,
				Value: fmt.Sprintf("LEN(%s)", args[0])}
		default:
			return newError(fmt.Sprintf("Invalid Type. %s %s", arg.Type(), arg.String()))
		}

	},
	"date": func(ir *IR, args ...object.Object) object.Object {
		if len(args) != 1 {
			return newError(fmt.Sprintf("wrong number of arguments. got=%d, want=1", len(args)))
		}

		arg := args[0]

		switch {
		case arg.Type() == object.STRING_OBJ:
			return &object.Expression{ExpressionType: object.DATE_OBJ,
				Value: fmt.Sprintf("CONVERT(date, %s, 23)", args[0])}
		default:
			return newError(fmt.Sprintf("Invalid Type. %s %s", arg.Type(), arg.String()))
		}
	},
	"exclude": func(ir *IR, args ...object.Object) object.Object {
		if len(args) != 1 {
			return newError(fmt.Sprintf("wrong number of arguments. got=%d, want=1", len(args)))
		}

		arg := args[0]

		switch arg.(type) {
		case *object.FieldCall:
			ir.currentSelectStatement.Exclude = true
			return arg
		default:
			return newError(fmt.Sprintf("Invalid function call for exclude. %s %s", arg.Type(), arg.String()))
		}
	},
}
