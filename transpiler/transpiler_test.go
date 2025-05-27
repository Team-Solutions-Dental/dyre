package transpiler

import (
	"testing"

	"github.com/vamuscari/dyre/endpoint"
	"github.com/vamuscari/dyre/lexer"
	"github.com/vamuscari/dyre/parser"
)

func TestEvalQueries(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"int:,", "SELECT Test.int FROM dbo.Test"},
		{"string: @ == 'Hello',", "SELECT Test.string FROM dbo.Test WHERE (Test.string = 'Hello')"},
		{"bool: @ == FALSE,", "SELECT Test.bool FROM dbo.Test WHERE (Test.bool = 0)"},
		{"int: > 5,", "SELECT Test.int FROM dbo.Test WHERE (Test.int > 5)"},
		{"int: > 5 || < 10,", "SELECT Test.int FROM dbo.Test WHERE ((Test.int > 5) OR (Test.int < 10))"},
		{"int: @ == 5,string: @ != null,", "SELECT Test.int, Test.string FROM dbo.Test WHERE (Test.int = 5) AND (Test.string != NULL)"},
		{"int: @ == 5,string: @ != NULL,", "SELECT Test.int, Test.string FROM dbo.Test WHERE (Test.int = 5) AND (Test.string != NULL)"},
		{"string: len(@) > 5,", "SELECT Test.string FROM dbo.Test WHERE (LEN(Test.string) > 5)"},
		{"date: @ == date('01/02/2023'),", "SELECT Test.date FROM dbo.Test WHERE (Test.date = CONVERT(date, '01/02/2023'))"},
		{"int: > 5", "SELECT Test.int FROM dbo.Test WHERE (Test.int > 5)"},
	}

	for _, tt := range tests {
		evalualted, err := testEval(tt.input)
		if err != nil {
			for _, e := range err {
				t.Errorf("Query test error. [%s] %s\n", tt.input, e)
			}
		}
		if evalualted != tt.expected {
			t.Errorf("Query failed. [%s]\n%s \n%s\n ", tt.input, evalualted, tt.expected)
		}
	}
}

func testEval(input string) (string, []string) {
	l := lexer.New(input)
	p := parser.New(l)
	q := p.ParseQuery()
	service := &endpoint.Service{
		Endpoints: map[string]*endpoint.Endpoint{
			"Test": {
				Name:       "Test",
				TableName:  "Test",
				SchemaName: "dbo",
				FieldNames: []string{"int", "string", "bool", "date"},
				Fields: map[string]endpoint.Field{
					"int":    {Name: "int", DefaultField: false},
					"string": {Name: "string", DefaultField: false},
					"bool":   {Name: "bool", DefaultField: false},
					"date":   {Name: "date", DefaultField: false},
				},
			},
		},
	}
	ir := &IR{
		Endpoint: service.Endpoints["Test"],
		Ast:      q,
	}

	ir.Eval()

	return ir.BuildSQLQuery(), ir.Errors
}

// func TestEvalIntegerExpression(t *testing.T) {
// 	integer := "5"
// 	field := &endpoint.Field{
// 		Name:         "int",
// 		TypeName:     "integer",
// 		SqlType:      "int",
// 		DefaultField: false,
// 		SqlSelect:    "int",
// 	}
//
// 	exp := testEvalExpression(integer, field)
// 	if exp != integer {
// 		t.Errorf("Integer is incorrect. got=%s, want=%s", exp, integer)
// 	}
// }
//
// func testEvalExpression(input string, field *endpoint.Field) string {
// 	l := lexer.New(input)
// 	p := parser.New(l)
// 	q := p.ParseQuery()
// 	req := &request.Request{
// 		Endpoint: &endpoint.Endpoint{
// 			Name:        "Test",
// 			RequestType: "GET",
// 			FieldNames:  []string{"int", "string", "boolean"},
// 			Fields: map[string]endpoint.Field{
// 				"int": {Name: "int", TypeName: "integer", SqlType: "integer", DefaultField: false, SqlSelect: "int"},
// 				// "string":  {"string", "string", "NVARCHAR(MAX)", false, "string"},
// 				// "boolean": {"boolean", "boolean", "bit", false, "boolean"},
// 			},
// 			TableName: "dbo.test",
// 		},
// 	}
//
// 	return evalExpression(q, req, field)
// }
