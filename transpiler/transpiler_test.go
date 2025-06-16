package transpiler

import (
	"testing"

	"github.com/vamuscari/dyre/endpoint"
	"github.com/vamuscari/dyre/object"
)

func TestEvalQueries(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"int:", "SELECT Test.[int] FROM dbo.Test"},                                                                                 // Basic Request
		{"int:string:bool:", "SELECT Test.[int], Test.[string], Test.[bool] FROM dbo.Test"},                                         // Chain Column
		{"int:;string:;int:", "SELECT Test.[string], Test.[int] FROM dbo.Test"},                                                     // Reorder
		{"string: @ == 'Hello'", "SELECT Test.[string] FROM dbo.Test WHERE (Test.[string] = 'Hello')"},                              // @ reference call
		{"bool: @ == FALSE", "SELECT Test.[bool] FROM dbo.Test WHERE (Test.[bool] = 0)"},                                            // Boolean Comparison
		{"int: int: > 5", "SELECT Test.[int] FROM dbo.Test WHERE (Test.[int] > 5)"},                                                 // Integer Comparison
		{"int: > 5 OR < 10", "SELECT Test.[int] FROM dbo.Test WHERE ((Test.[int] > 6) OR (Test.[int] < 10))"},                       // OR Statement
		{"int: > 5 AND < 10", "SELECT Test.[int] FROM dbo.Test WHERE ((Test.[int] > 5) AND (Test.[int] < 10))"},                     // AND Statement
		{"date: @ == date('01/02/2023')", "SELECT Test.[date] FROM dbo.Test WHERE (Test.[date] = CONVERT(date, '01/02/2023', 23))"}, // Function Call
		{"bool: exclude(@); == true;string:", "SELECT Test.[string] FROM dbo.Test WHERE (Test.[bool] = 1)"},                         // Exclude
		{"string: alias('str');", "SELECT (Test.[string]) AS 'str' FROM dbo.Test"},                                                  // alias

		//{"alias('test');bool: exclude(@); == true;string:", "SELECT Test.[string] FROM dbo.Test WHERE (Test.[bool] = 1)"},           // TableFunction
		// {"int: @ == 5 string: @ != NULL", "SELECT Test.[int], Test.[string] FROM dbo.Test WHERE (Test.[int] = 5) AND (Test.[string] != NULL)"}, // NULL Comparison
	}

	for _, tt := range tests {
		ir, err := testNew(tt.input)
		if err != nil {
			t.Errorf("Query test error. [%s] %s\n", tt.input, err.Error())
		}
		sql_statement, err := ir.EvaluateQuery()

		if err != nil {
			t.Errorf("Query test error. [%s] %s\n", tt.input, err.Error())
		}

		if sql_statement != tt.expected {
			t.Errorf("Query failed. [%s]\n%s \n%s\n ", tt.input, sql_statement, tt.expected)
		}
	}
}

func TestLimit(t *testing.T) {
	tests := []struct {
		input    string
		limit    int
		expected string
	}{
		{"int:", 100, "SELECT TOP 100 Test.[int] FROM dbo.Test"},
		{"int:", -1, "SELECT Test.[int] FROM dbo.Test"},
	}

	for _, tt := range tests {
		evalualted_ir, err := testNew(tt.input)
		if err != nil {
			t.Errorf("Query test error. [%s] %s\n", tt.input, err.Error())
		}
		evalualted_ir.LIMIT(tt.limit)
		sql_statement, err := evalualted_ir.EvaluateQuery()

		if err != nil {
			t.Errorf("Query test error. [%s] %s\n", tt.input, err.Error())
		}

		if sql_statement != tt.expected {
			t.Errorf("Query failed. [%s]\n%s \n%s\n ", tt.input, sql_statement, tt.expected)
		}
	}

}

func testNew(input string) (*PrimaryIR, error) {
	service := &endpoint.Service{
		Endpoints: map[string]*endpoint.Endpoint{
			"Test": {
				Name:       "Test",
				TableName:  "Test",
				SchemaName: "dbo",
				FieldNames: []string{"int", "string", "bool", "date"},
				Fields: map[string]endpoint.Field{
					"int":    {Name: "int", FieldType: object.INTEGER_OBJ, Nullable: true},
					"string": {Name: "string", FieldType: object.STRING_OBJ, Nullable: true},
					"bool":   {Name: "bool", FieldType: object.BOOLEAN_OBJ, Nullable: true},
					"date":   {Name: "date", FieldType: object.DATE_OBJ, Nullable: true},
				},
			},
		},
	}
	return New(input, service.Endpoints["Test"])
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
