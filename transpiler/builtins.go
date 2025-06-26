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
	//cast(expression, to)
	"cast": func(ir *IR, local *objectRef.LocalReferences, args ...object.Object) object.Object {
		if len(args) != 2 {
			return newError("wrong number of arguments. got=%d, want=2", len(args))
		}

		expression := args[0]

		if args[1].Type() != objectType.STRING {
			return newError("Invalid cast to identity type, got=%s, want=STRING", args[1].Type())
		}

		castTo, ok := args[1].(*object.String)
		if !ok {
			return newError("Invalid cast to identity type, got=%s, want=STRING", args[1].Type())
		}

		return &object.Expression{ExpressionType: objectType.EXPRESSION,
			Value: fmt.Sprintf("CAST( %s AS %s )", expression, castTo.Value)}

	},
	"timezone": func(ir *IR, local *objectRef.LocalReferences, args ...object.Object) object.Object {
		if len(args) != 2 {
			return newError("wrong number of arguments. got=%d, want=2", len(args))
		}

		expression := args[0]

		zone, ok := args[1].(*object.String)
		if !ok {
			return newError("Invalid Argument Type. %s %s", args[1].Type(), args[1].String())
		}

		return &object.Expression{ExpressionType: objectType.DATE,
			Value: fmt.Sprintf("%s AT TIME ZONE %s", expression, zone.String())}
	},
	"datepart": func(ir *IR, local *objectRef.LocalReferences, args ...object.Object) object.Object {
		if len(args) != 2 {
			return newError("wrong number of arguments. got=%d, want=2", len(args))
		}

		datepart, ok := args[0].(*object.String)
		if !ok {
			return newError("Invalid Argument Type. %s %s", datepart.Type(), datepart.String())
		}

		return &object.Expression{ExpressionType: objectType.DATE,
			Value: fmt.Sprintf("DATEPART(%s, %s)", datepart.Value, args[1])}
	},
	"dateadd": func(ir *IR, local *objectRef.LocalReferences, args ...object.Object) object.Object {
		if len(args) != 3 {
			return newError("wrong number of arguments. got=%d, want=3", len(args))
		}

		interval, ok := args[0].(*object.String)
		if !ok {
			return newError("Invalid Argument Type (Expect String). %s %s", args[0].Type(), args[0].String())
		}

		num, ok := args[1].(*object.Integer)
		if !ok {
			return newError("Invalid Argument Type (Expect Int). %s %s", args[1].Type(), args[1].String())
		}

		return &object.Expression{ExpressionType: objectType.DATE,
			Value: fmt.Sprintf("DATEADD(%s, %s, %s)", interval.Value, num.String(), args[2])}
	},
	"convert": func(ir *IR, local *objectRef.LocalReferences, args ...object.Object) object.Object {
		if len(args) < 2 {
			return newError("wrong number of arguments. got=%d, want=2-3", len(args))
		}

		if len(args) > 3 {
			return newError("wrong number of arguments. got=%d, want=2-3", len(args))
		}

		convert, ok := args[0].(*object.String)
		if !ok {
			return newError("Invalid Argument Type. %s %s", args[0].Type(), args[0].String())
		}

		if len(args) == 2 {
			return &object.Expression{ExpressionType: objectType.DATE,
				Value: fmt.Sprintf("CONVERT(%s, %s)", convert.Value, args[1])}
		}

		if len(args) == 3 {
			return &object.Expression{ExpressionType: objectType.DATE,
				Value: fmt.Sprintf("CONVERT(%s, %s, %s)", convert.Value, args[1], args[2])}
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
