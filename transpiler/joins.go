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

func (jt *joinType) ON(parent_on, on string) *joinStatement {
	return &joinStatement{Type: jt.Type, parentIR: jt.parentIR, endpoint: jt.endpoint, errors: jt.errors, parent_on: parent_on, on: on}
}

type joinStatement struct {
	Join
	Type      string
	parentIR  *IR
	endpoint  *endpoint.Endpoint
	errors    []error
	parent_on string
	on        string
	joinType  string
	ir        *IR
}

func (js *joinStatement) Errors() []error {
	return js.errors
}

func (js *joinStatement) Query(query string) (*IR, []error) {
	if js.errors != nil {
		return nil, js.errors
	}

	js.ir = New(query, js.endpoint)

	// append joined fields except on field
	js.parentIR.joinStatements = append(js.parentIR.joinStatements, js)

	// do this after on SQL Build
	// parent_contains := utils.Array_Contains(js.parentIR.FieldNames(), js.parent_on)
	// if !parent_contains {
	// 	js.errors = append(js.errors, errors.New(fmt.Sprintf("Parent query %s does not contain %s", js.parentIR.endpoint.Name, js.parent_on)))
	//
	// }
	//
	// join_contains := utils.Array_Contains(js.ir.FieldNames(), js.on)
	// if !join_contains {
	// 	js.errors = append(js.errors, errors.New(fmt.Sprintf("Join query %s does not contain %s", js.ir.endpoint.Name, js.on)))
	// }

	return js.ir, js.errors
}

func (js *joinStatement) appendSelectStatements() {
	parent_contains := utils.Array_Contains(js.parentIR.FieldNames(), js.parent_on)
	if !parent_contains {
		js.errors = append(js.errors, errors.New(fmt.Sprintf("Parent query %s does not contain %s", js.parentIR.endpoint.Name, js.parent_on)))
	}

	join_contains := utils.Array_Contains(js.ir.FieldNames(), js.on)
	if !join_contains {
		js.errors = append(js.errors, errors.New(fmt.Sprintf("Join query %s does not contain %s", js.ir.endpoint.Name, js.on)))
	}

	for _, ss := range js.ir.selectStatements {
		if (*ss.fieldName) != js.on {
			js.parentIR.selectStatements = append(js.parentIR.selectStatements, &selectStatement{fieldName: ss.fieldName, tableName: &js.ir.endpoint.TableName})
		}
	}
}

func (js *joinStatement) parentIrOn() string {
	return fmt.Sprintf("%s.%s", js.parentIR.endpoint.TableName, js.parent_on)
}

func (js *joinStatement) joinIrOn() string {
	return fmt.Sprintf("%s.%s", js.ir.endpoint.TableName, js.on)
}
