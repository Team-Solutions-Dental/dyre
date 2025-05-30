package object

import (
	"fmt"
)

type ObjectType string

const (
	INTEGER_OBJ    = "INTEGER"
	FLOAT_OBJ      = "FLOAT"
	BOOLEAN_OBJ    = "BOOLEAN"
	STRING_OBJ     = "STRING"
	DATE_OBJ       = "DATE"
	DATETIME_OBJ   = "DATETIME"
	EXPRESSION_OBJ = "EXPRESSION"
	STATEMENT_OBJ  = "STATEMENT"
	NULL_OBJ       = "NULL"
	ERROR_OBJ      = "ERROR"
	BUILTIN_OBJ    = "BUILTIN"
)

type Object interface {
	Type() ObjectType
	String() string
}

type Expression struct {
	ExpressionType ObjectType
	Value          string
}

func (e *Expression) Type() ObjectType { return e.ExpressionType }
func (e *Expression) String() string   { return e.Value }

type FieldCall struct {
	FieldType ObjectType
	Value     string
}

func (fc *FieldCall) Type() ObjectType { return fc.FieldType }
func (fc *FieldCall) String() string   { return fc.Value }

type Integer struct {
	Value int64
}

func (i *Integer) Type() ObjectType { return INTEGER_OBJ }
func (i *Integer) String() string   { return fmt.Sprintf("%d", i.Value) }

type Float struct {
	Value float64
}

func (i *Float) Type() ObjectType { return FLOAT_OBJ }
func (i *Float) String() string   { return fmt.Sprintf("%f", i.Value) }

type Boolean struct {
	Value bool
}

func (b *Boolean) Type() ObjectType { return BOOLEAN_OBJ }
func (b *Boolean) String() string {
	if b.Value == true {
		return "1"
	}
	return "0"
}

type BooleanExpression struct {
	Value string
}

func (be *BooleanExpression) Type() ObjectType { return BOOLEAN_OBJ }
func (be *BooleanExpression) String() string   { return be.Value }

type BuiltinFunction func(args ...Object) Object

type Null struct{}

func (n *Null) Type() ObjectType { return NULL_OBJ }
func (n *Null) String() string   { return "NULL" }

type Error struct {
	Message string
}

func (e *Error) Type() ObjectType { return ERROR_OBJ }
func (e *Error) String() string   { return "ERROR: " + e.Message }

type String struct {
	Value string
}

func (s *String) Type() ObjectType { return STRING_OBJ }
func (s *String) String() string   { return fmt.Sprintf("'%s'", s.Value) }

type Builtin struct {
	Fn BuiltinFunction
}

func (b *Builtin) Type() ObjectType { return BUILTIN_OBJ }
func (b *Builtin) String() string   { return "Builtin function " }
