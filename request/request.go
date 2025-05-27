package request

import (
	"github.com/vamuscari/dyre/ast"
	"github.com/vamuscari/dyre/endpoint"
	"github.com/vamuscari/dyre/lexer"
	"github.com/vamuscari/dyre/parser"
	"github.com/vamuscari/dyre/transpiler"
)

func New(endpoint *endpoint.Endpoint) *Request {
	return &Request{Endpoint: endpoint}
}

type Request struct {
	Endpoint *endpoint.Endpoint
	Ast      *ast.QueryStatements
	IR       *transpiler.IR
}

func (r *Request) EvaluateQuery(input string) {
	l := lexer.New(input)
	p := parser.New(l)
	r.Ast = p.ParseQuery()
}

func (r *Request) BuildSQL() string {
	return r.IR.BuildSQLQuery()
}
