package transpiler

import (
	"github.com/vamuscari/dyre/endpoint"
	"github.com/vamuscari/dyre/sql"
)

type joinType struct {
	joinType string
	parentIR *IR
	name     string
}

func (jt *joinType) ON(parent_on, on string) *joinIR {
	return &joinIR{joinType: jt.joinType, parentIR: jt.parentIR, parent_on: parent_on, child_on: on, name: jt.name}
}

type joinIR struct {
	name      string
	parentIR  *IR
	endpoint  *endpoint.Endpoint
	errors    []error
	parent_on string
	child_on  string
	joinType  string
	child_ir  *IR
}

func (js *joinIR) Query(query string) (*IR, error) {

	ep, err := js.endpoint.Service.GetEndpoint(js.name)
	if err != nil {
		return nil, err
	}

	js.endpoint = ep

	js.child_ir = New(query, js.endpoint)

	// append joined fields except on field
	js.parentIR.joins = append(js.parentIR.joins, js)

	joinStmnt := &sql.JoinStatement{
		Parent_Query: js.parentIR.sql,
		Child_Query:  js.child_ir.sql,
		Parent_On:    js.parent_on,
		Child_On:     js.child_on,
		Alias:        js.name,
	}

	js.parentIR.sql.JoinStatements = append(js.parentIR.sql.JoinStatements, joinStmnt)

	return js.child_ir, nil
}
