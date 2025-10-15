package transpiler

import (
	"fmt"
	"github.com/vamuscari/dyre/endpoint"
	"github.com/vamuscari/dyre/sql"
	"strings"
)

func (ir *IR) AUTOJOIN(joinPrefix string, joinEndpointName string) (*joinIR, error) {
	endpointJoin, ok := ir.endpoint.Joins[joinEndpointName]
	if !ok {
		return nil, fmt.Errorf("Auto Join, Endpoint %s does not contain join %s", ir.endpoint.Name, joinEndpointName)
	}

	joinPrefix, err := JoinPrefixEval(joinPrefix)
	if err != nil {
		return nil, fmt.Errorf("Auto join %s, %w", joinEndpointName, err)
	}

	autojoin := &joinIR{
		joinType: joinPrefix,
		parentIR: ir,
		name:     joinEndpointName,
		parentOn: endpointJoin.Parent_ON,
		childOn:  endpointJoin.Child_ON,
		endpoint: endpointJoin.ChildEndpoint(),
		alias:    joinEndpointName,
	}

	return autojoin, nil

}

func (ir *IR) INNERJOIN(req string) *joinType {
	join := &joinType{joinType: "INNER", parentIR: ir, name: req}

	return join

}

func (ir *IR) LEFTJOIN(req string) *joinType {
	join := &joinType{joinType: "LEFT", parentIR: ir, name: req}

	return join
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
	return &joinIR{
		joinType: jt.joinType,
		parentIR: jt.parentIR,
		parentOn: parent_on,
		childOn:  on,
		name:     jt.name,
		alias:    jt.name,
	}
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

	js.childIR, err = newSubIRWithSecurity(query, js.endpoint, js.parentIR.securityChecker)
	if err != nil {
		return nil, err
	}

	js.parentIR.joins = append(js.parentIR.joins, js)

	joinStmnt := &sql.JoinStatement{
		JoinType:     &js.joinType,
		Parent_Query: js.parentIR.sql,
		Child_Query:  js.childIR.sql,
		Parent_On:    &js.parentOn,
		Child_On:     &js.childOn,
		Alias:        &js.name,
	}

	// if js.parentIR.sql.SelectStatements == nil {
	//
	// } else {
	// 	js.parentIR.sql.JoinStatements = append(js.parentIR.sql.JoinStatements, joinStmnt)
	// }

	if js.parentIR.sql.JoinStatements == nil {
		js.parentIR.sql.JoinStatements = []*sql.JoinStatement{joinStmnt}
	} else {
		js.parentIR.sql.JoinStatements = append(js.parentIR.sql.JoinStatements, joinStmnt)
	}

	return js.childIR, nil
}

// Make sure that the select statement needed for joining is on the child table
func (js *joinIR) Check() error {
	childLoc := js.childIR.sql.SelectStatementLocation(js.childOn)
	if childLoc < 0 {
		childField := js.childIR.endpoint.Fields[js.childOn]
		ss := childField.SelectStatement()
		ss.Query = js.childIR.sql
		js.childIR.sql.SelectStatements = append(js.childIR.sql.SelectStatements, ss)
	}

	parentLoc := js.parentIR.sql.SelectStatementLocation(js.parentOn)
	if parentLoc < 0 {
		_, ok := js.parentIR.endpoint.Fields[js.parentOn]
		if !ok {
			return fmt.Errorf("No field '%s' found to join on endpoint '%s'", js.parentOn, js.parentIR.endpoint.Name)
		}
	}

	return nil
}

func JoinPrefixEval(input string) (string, error) {
	input = strings.ToUpper(input)
	switch input {
	case "INNER":
		return "INNER", nil
	case "LEFT":
		return "LEFT", nil
	case "RIGHT":
		return "RIGHT", nil
	case "FULL":
		return "FULL", nil
	default:
		return "INNER", fmt.Errorf("Invalid Join Prefix %s", input)
	}

}
