package ast

import (
	"bytes"
	"github.com/vamuscari/dyre/token"
	"strings"
)

type Node interface {
	TokenLiteral() string
	String() string
}

type Statement interface {
	Node
	statementNode()
}

type Expression interface {
	Node
	expressionNode()
}

// All statements made in a request.
// Seperated by ; or columns.
type RequestStatements struct {
	Statements []Statement
}

func (p *RequestStatements) TokenLiteral() string {
	if len(p.Statements) > 0 {
		return p.Statements[0].TokenLiteral()
	} else {
		return ""
	}
}

func (p *RequestStatements) String() string {
	var out bytes.Buffer

	for _, c := range p.Statements {
		out.WriteString(c.String())
	}

	return out.String()
}

// The call of a column into the output
// Ex. FirstName:
type ColumnLiteral struct {
	Token token.Token
}

func (cs *ColumnLiteral) statementNode()       {}
func (cs *ColumnLiteral) TokenLiteral() string { return cs.Token.Literal }
func (cs *ColumnLiteral) String() string {
	var out bytes.Buffer
	out.WriteString(cs.TokenLiteral() + ": ")
	return out.String()
}

// The call of a Column Function into the output
// Ex. AS('isTrue', @('true'))
// Ex. AS('year', datepart('year', @('Created')))
type ColumnFunction struct {
	Token     token.Token
	Fn        string
	Arguments []Expression
}

func (cf *ColumnFunction) statementNode()       {}
func (cf *ColumnFunction) TokenLiteral() string { return cf.Token.Literal }
func (cf *ColumnFunction) String() string {
	var out bytes.Buffer

	out.WriteString(cf.TokenLiteral())
	out.WriteString("(")

	args := []string{}
	for _, a := range cf.Arguments {
		args = append(args, a.String())
	}

	out.WriteString(strings.Join(args, ", "))

	out.WriteString(")")
	out.WriteString(": ")
	return out.String()
}

// The call of a Group Function into the output
// Ex. GROUP('Active'): AVG('AvgSales', @('Sales')):
type GroupFunction struct {
	Token     token.Token
	Fn        string
	Arguments []Expression
}

func (cs *GroupFunction) statementNode()       {}
func (cs *GroupFunction) TokenLiteral() string { return cs.Token.Literal }
func (cs *GroupFunction) String() string {
	var out bytes.Buffer

	out.WriteString(cs.TokenLiteral())
	out.WriteString("(")

	args := []string{}
	for _, a := range cs.Arguments {
		args = append(args, a.String())
	}

	out.WriteString(strings.Join(args, ", "))

	out.WriteString(")")
	out.WriteString(": ")
	return out.String()
}

type ExpressionStatement struct {
	Token      token.Token
	Expression Expression
}

func (es *ExpressionStatement) statementNode()       {}
func (es *ExpressionStatement) TokenLiteral() string { return es.Token.Literal }
func (es *ExpressionStatement) String() string {
	if es.Expression != nil {
		return es.Expression.String()
	}
	return ""
}

type Identifier struct {
	Token token.Token // the IDENT token
	Value string
}

func (i *Identifier) expressionNode()      {}
func (i *Identifier) TokenLiteral() string { return i.Token.Literal }
func (i *Identifier) String() string       { return i.Value }

type IntegerLiteral struct {
	Token token.Token
	Value int64
}

func (il *IntegerLiteral) expressionNode()      {}
func (il *IntegerLiteral) TokenLiteral() string { return il.Token.Literal }
func (il *IntegerLiteral) String() string       { return il.Token.Literal }

type OrderExpression struct {
	Token     token.Token
	Ascending bool
}

func (ol *OrderExpression) expressionNode()      {}
func (ol *OrderExpression) TokenLiteral() string { return ol.Token.Literal }
func (ol *OrderExpression) String() string       { return ol.Token.Literal }

type NullLiteral struct {
	Token token.Token
}

func (nl *NullLiteral) expressionNode()      {}
func (nl *NullLiteral) TokenLiteral() string { return nl.Token.Literal }
func (nl *NullLiteral) String() string       { return "NULL" }

type PrefixExpression struct {
	Token    token.Token
	Operator string
	Right    Expression
}

func (pe *PrefixExpression) expressionNode()      {}
func (pe *PrefixExpression) TokenLiteral() string { return pe.Token.Literal }
func (pe *PrefixExpression) String() string {
	var out bytes.Buffer

	out.WriteString("(")
	out.WriteString(pe.Operator)
	out.WriteString(pe.Right.String())
	out.WriteString(")")

	return out.String()
}

type InfixExpression struct {
	Token    token.Token
	Left     Expression
	Operator string
	Right    Expression
}

func (ie *InfixExpression) expressionNode()      {}
func (ie *InfixExpression) TokenLiteral() string { return ie.Token.Literal }
func (ie *InfixExpression) String() string {
	var out bytes.Buffer

	out.WriteString("(")
	out.WriteString(ie.Left.String())
	out.WriteString(" " + ie.Operator + " ")
	out.WriteString(ie.Right.String())
	out.WriteString(")")

	return out.String()
}

type Boolean struct {
	Token token.Token
	Value bool
}

func (b *Boolean) expressionNode()      {}
func (b *Boolean) TokenLiteral() string { return b.Token.Literal }
func (b *Boolean) String() string       { return b.Token.Literal }

type CallExpression struct {
	Token     token.Token
	Function  Expression
	Arguments []Expression
}

func (ce *CallExpression) expressionNode()      {}
func (ce *CallExpression) TokenLiteral() string { return ce.Token.Literal }
func (ce *CallExpression) String() string {
	var out bytes.Buffer

	args := []string{}
	for _, a := range ce.Arguments {
		args = append(args, a.String())
	}

	out.WriteString(ce.Function.String())
	out.WriteString("(")
	out.WriteString(strings.Join(args, ", "))
	out.WriteString(")")

	return out.String()
}

type StringLiteral struct {
	Token token.Token
	Value string
}

func (sl *StringLiteral) expressionNode()      {}
func (sl *StringLiteral) TokenLiteral() string { return sl.Token.Literal }
func (sl *StringLiteral) String() string {
	return "'" + sl.Token.Literal + "'"
}

// @ token
// Can reference the last Column or be used as a function to call a column
type Reference struct {
	Token    token.Token
	Argument Expression
}

func (r *Reference) expressionNode()      {}
func (r *Reference) TokenLiteral() string { return r.Token.Literal }
func (r *Reference) String() string {
	var out bytes.Buffer

	out.WriteString("@")
	if r.Argument == nil {
		return out.String()
	}

	out.WriteString("(")
	out.WriteString(r.Argument.String())
	out.WriteString(")")

	return out.String()
}
