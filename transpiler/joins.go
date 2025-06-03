package transpiler

import (
	"github.com/vamuscari/dyre/endpoint"
	"github.com/vamuscari/dyre/sql"
)

func (ir *IR) INNERJOIN(req string) *joinType {
	join := &joinType{joinType: "INNER", parentIR: ir, name: req}

	return join

}

func (ir *IR) LEFTJOIN(req string) *joinType {
	join := &joinType{joinType: "LEFT", parentIR: ir, name: req}

	return join
}

func (ir *IR) FieldNames() []string {
	return ir.sql.SelectNameList()
}

func (ir *IR) LIMIT(input int) *IR {
	ir.sql.Limit = &input
	return ir
}

type joinType struct {
	joinType string
	parentIR *IR
	name     string
}

func (jt *joinType) ON(parent_on, on string) *joinIR {
	return &joinIR{joinType: jt.joinType, parentIR: jt.parentIR, parentOn: parent_on, childOn: on, name: jt.name, alias: jt.name}
}

type joinIR struct {
	name     string
	alias    string
	parentIR *IR
	childIR  *SubIR
	endpoint *endpoint.Endpoint
	parentOn string
	childOn  string
	joinType string
}

func (js *joinIR) Query(query string) (*SubIR, error) {

	ep, err := js.parentIR.endpoint.Service.GetEndpoint(js.name)
	if err != nil {
		return nil, err
	}

	js.endpoint = ep

	js.childIR, err = newSubIR(query, js.endpoint)
	if err != nil {
		return nil, err
	}

	js.childIR.sql.Depth = js.parentIR.sql.Depth + 1

	// append joined fields except on field
	js.parentIR.joins = append(js.parentIR.joins, js)

	joinStmnt := &sql.JoinStatement{
		JoinType:     &js.joinType,
		Parent_Query: js.parentIR.sql,
		Child_Query:  js.childIR.sql,
		Parent_On:    &js.parentOn,
		Child_On:     &js.childOn,
		Alias:        &js.name,
	}

	if js.parentIR.sql.JoinStatements == nil {
		js.parentIR.sql.JoinStatements = []*sql.JoinStatement{joinStmnt}
	} else {
		js.parentIR.sql.JoinStatements = append(js.parentIR.sql.JoinStatements, joinStmnt)
	}

	return js.childIR, nil
}
