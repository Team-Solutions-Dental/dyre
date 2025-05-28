package transpiler

import (
	"testing"

	"github.com/vamuscari/dyre/endpoint"
	"github.com/vamuscari/dyre/lexer"
	"github.com/vamuscari/dyre/parser"
)

func TestEvalJoins(t *testing.T) {
	tests := []struct {
		input_parent string
		input_join   string
		on           string
		expected     string
	}{
		{"int:,string:,", "int:,bool:,", "int", "SELECT Parent.int, Parent.string, Join.bool FROM Parent INNER JOIN ( SELECT Join.int, Join.bool FROM Join ) AS Join ON Parent.int = Join.int"},
	}

	for _, tt := range tests {
		parent_ir, errs := testParentEval(tt.input_parent)
		if errs != nil {
			for _, e := range errs {
				t.Errorf("Query test error. [%s] %s\n", tt.input_parent, e)
			}
		}

		eerrs := parent_ir.INNERJOIN("Join").ON(tt.on).Query(tt.input_join)
		if eerrs != nil {
			for _, e := range eerrs {
				t.Errorf("Query test error. [%s] %v\n", tt.input_parent, e)
			}
		}

		evaluated := parent_ir.BuildSQLQuery()

		if evaluated != tt.expected {
			t.Errorf("Query failed. [%s] [%s]\n%s \n%s\n ", tt.input_parent, tt.input_join, evaluated, tt.expected)
		}

	}
}

func testParentEval(input string) (*IR, []string) {
	l := lexer.New(input)
	p := parser.New(l)
	q := p.ParseQuery()
	var service *endpoint.Service = &endpoint.Service{Endpoints: map[string]*endpoint.Endpoint{}}

	parent := endpoint.Endpoint{
		Service:    service,
		Name:       "Parent",
		TableName:  "Parent",
		FieldNames: []string{"int", "string", "bool", "date"},
		Fields: map[string]endpoint.Field{
			"int":    {Name: "int", DefaultField: false},
			"string": {Name: "string", DefaultField: false},
			"date":   {Name: "date", DefaultField: false},
		},
	}

	join := endpoint.Endpoint{
		Service:    service,
		Name:       "Join",
		TableName:  "Join",
		FieldNames: []string{"int", "string", "bool", "date"},
		Fields: map[string]endpoint.Field{
			"int":    {Name: "int", DefaultField: false},
			"string": {Name: "string", DefaultField: false},
			"bool":   {Name: "bool", DefaultField: false},
			"date":   {Name: "date", DefaultField: false},
		},
	}

	service.Endpoints["Parent"] = &parent
	service.Endpoints["Join"] = &join

	ir := &IR{
		endpoint: service.Endpoints["Parent"],
		ast:      q,
	}

	ir.eval()

	return ir, ir.Errors
}
