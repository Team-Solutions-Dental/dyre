package transpiler

import (
	"errors"
	"fmt"

	"github.com/vamuscari/dyre/endpoint"
	"github.com/vamuscari/dyre/utils"
)

type Join interface {
	Errors()
}

type joinType struct {
	Join
	Type     string
	parentIR *IR
	endpoint *endpoint.Endpoint
	errors   []error
}

func (jt *joinType) Errors() []error {
	return jt.errors
}

func (jt *joinType) ON(field string) *joinStatement {
	return &joinStatement{Type: jt.Type, parentIR: jt.parentIR, endpoint: jt.endpoint, errors: jt.errors, on: field}
}

type joinStatement struct {
	Join
	Type     string
	parentIR *IR
	endpoint *endpoint.Endpoint
	errors   []error
	on       string
	joinType string
	ir       *IR
}

func (js *joinStatement) Errors() []error {
	return js.errors
}

func (js *joinStatement) Query(query string) []error {
	if js.errors != nil {
		return js.Errors()
	}

	js.ir = New(query, js.endpoint)

	// do this after on SQL Build
	parent_contains := utils.Array_Contains(js.parentIR.FieldNames(), js.on)
	if !parent_contains {
		js.errors = append(js.errors, errors.New(fmt.Sprintf("Parent query %s does not contain %s", js.parentIR.endpoint.Name, js.on)))

	}

	join_contains := utils.Array_Contains(js.ir.FieldNames(), js.on)
	if !join_contains {
		js.errors = append(js.errors, errors.New(fmt.Sprintf("Join query %s does not contain %s", js.ir.endpoint.Name, js.on)))
	}

	if js.errors != nil {
		return js.Errors()
	}

	for _, ss := range js.ir.selectStatements {
		fmt.Println(*ss.fieldName)
		if (*ss.fieldName) != js.on {
			js.parentIR.selectStatements = append(js.parentIR.selectStatements, ss)
		}
	}

	js.parentIR.joinStatements = append(js.parentIR.joinStatements, js)

	return nil
}

func (js *joinStatement) parentIrOn() string {
	return fmt.Sprintf("%s.%s", js.parentIR.endpoint.TableName, js.on)
}

func (js *joinStatement) joinIrOn() string {
	return fmt.Sprintf("%s.%s", js.ir.endpoint.TableName, js.on)
}
