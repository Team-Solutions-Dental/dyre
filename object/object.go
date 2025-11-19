package object

import (
	"fmt"

	"github.com/Team-Solutions-Dental/dyre/object/objectType"
)

type Object interface {
	Type() objectType.Type
	String() string
	Nullable() bool
}

type Expression struct {
	ExpressionType objectType.Type
	Value          string
	HasNull        bool
}

func (e *Expression) Type() objectType.Type { return e.ExpressionType }
func (e *Expression) Nullable() bool        { return e.HasNull }
func (e *Expression) String() string        { return e.Value }

type Integer struct {
	Value int64
}

func (i *Integer) Type() objectType.Type { return objectType.INTEGER }
func (i *Integer) Nullable() bool        { return false }
func (i *Integer) String() string        { return fmt.Sprintf("%d", i.Value) }

type Float struct {
	Value float64
}

func (f *Float) Type() objectType.Type { return objectType.FLOAT }
func (f *Float) Nullable() bool        { return false }
func (f *Float) String() string        { return fmt.Sprintf("%f", f.Value) }

type Boolean struct {
	Value bool
}

func (b *Boolean) Type() objectType.Type { return objectType.BOOLEAN }
func (b *Boolean) Nullable() bool        { return false }
func (b *Boolean) String() string {
	if b.Value == true {
		return "1"
	}
	return "0"
}

type Null struct{}

func (n *Null) Type() objectType.Type { return objectType.NULL }
func (n *Null) Nullable() bool        { return false }
func (n *Null) String() string        { return "NULL" }

type Error struct {
	Message string
}

func (e *Error) Type() objectType.Type { return objectType.ERROR }
func (e *Error) Nullable() bool        { return false }
func (e *Error) String() string        { return "ERROR: " + e.Message }

type String struct {
	Value string
}

func (s *String) Type() objectType.Type { return objectType.STRING }
func (s *String) Nullable() bool        { return false }
func (s *String) String() string        { return fmt.Sprintf("'%s'", s.Value) }

type BuiltinFunction func(args ...Object) Object

type Builtin struct {
	Fn BuiltinFunction
}

func (b *Builtin) Type() objectType.Type { return objectType.BUILTIN }
func (b *Builtin) Nullable() bool        { return false }
func (b *Builtin) String() string        { return "Builtin function " }

func CastType[T Object](obj Object) (T, *Error) {
	cast, ok := obj.(T)
	if !ok || cast.Type() != obj.Type() {
		return cast, &Error{Message: fmt.Sprintf("ERROR: Failed to cast %s", obj.Type())}
	}
	return cast, nil
}
