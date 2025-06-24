package transpiler

import (
	"testing"

	"github.com/vamuscari/dyre/endpoint"
	"github.com/vamuscari/dyre/object/objectType"
)

func TestEvalQueries(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"int:", "SELECT Test.[int] FROM dbo.Test"},                                                                                          // Basic Request
		{"int:string:bool:", "SELECT Test.[int], Test.[string], Test.[bool] FROM dbo.Test"},                                                  // Chain Column
		{"int:;string:;int:", "SELECT Test.[string], Test.[int] FROM dbo.Test"},                                                              // Reorder
		{"string: @ == 'Hello'", "SELECT Test.[string] FROM dbo.Test WHERE (Test.[string] = 'Hello')"},                                       // @ reference call
		{"bool: @ == FALSE", "SELECT Test.[bool] FROM dbo.Test WHERE (Test.[bool] = 0)"},                                                     // Boolean Comparison
		{"int: int: > 5", "SELECT Test.[int] FROM dbo.Test WHERE (Test.[int] > 5)"},                                                          // Integer Comparison
		{"int: > 5 OR < 10", "SELECT Test.[int] FROM dbo.Test WHERE ((Test.[int] > 5) OR (Test.[int] < 10))"},                                // OR Statement
		{"int: > 5 AND < 10", "SELECT Test.[int] FROM dbo.Test WHERE ((Test.[int] > 5) AND (Test.[int] < 10))"},                              // AND Statement
		{"date: @ == date('01/02/2023')", "SELECT Test.[date] FROM dbo.Test WHERE (Test.[date] = CONVERT(date, '01/02/2023', 23))"},          // Function Call
		{"int2:", "SELECT Test.[int2] FROM dbo.Test"},                                                                                        // AlphaNumeric Column
		{"bool: == NULL;", "SELECT Test.[bool] FROM dbo.Test WHERE (Test.[bool] IS NULL)"},                                                   // IS NULL
		{"bool: != NULL;", "SELECT Test.[bool] FROM dbo.Test WHERE (Test.[bool] IS NOT NULL)"},                                               // IS NOT NULL
		{"@('string') != NULL; bool: == false;", "SELECT Test.[bool] FROM dbo.Test WHERE (Test.[string] IS NOT NULL) AND (Test.[bool] = 0)"}, // @() reference function
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
				FieldNames: []string{"int", "int2", "string", "bool", "date"},
				Fields: map[string]endpoint.Field{
					"int":    {Name: "int", FieldType: objectType.INTEGER, Nullable: true},
					"int2":   {Name: "int2", FieldType: objectType.INTEGER, Nullable: true},
					"string": {Name: "string", FieldType: objectType.STRING, Nullable: true},
					"bool":   {Name: "bool", FieldType: objectType.BOOLEAN, Nullable: true},
					"date":   {Name: "date", FieldType: objectType.DATE, Nullable: true},
				},
			},
		},
	}
	return New(input, service.Endpoints["Test"])
}

func TestComparisonExpressions(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"a: > 5", "SELECT X.[a] FROM dbo.X WHERE (X.[a] > 5)"},   // GT >
		{"a: < 5", "SELECT X.[a] FROM dbo.X WHERE (X.[a] < 5)"},   // LT <
		{"a: >= 5", "SELECT X.[a] FROM dbo.X WHERE (X.[a] >= 5)"}, // GTE >=
		{"a: <= 5", "SELECT X.[a] FROM dbo.X WHERE (X.[a] <= 5)"}, // LTE <=
		{"a: == 5", "SELECT X.[a] FROM dbo.X WHERE (X.[a] = 5)"},  // EQ ==
		{"a: != 5", "SELECT X.[a] FROM dbo.X WHERE (X.[a] != 5)"}, // NOT_EQ !=
	}

	for _, tt := range tests {
		ir, err := testNewXYZ(tt.input)
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

func testNewXYZ(input string) (*PrimaryIR, error) {
	var service *endpoint.Service = &endpoint.Service{Endpoints: map[string]*endpoint.Endpoint{}}
	service.EndpointNames = []string{"X", "XY", "YZ"}

	x := &endpoint.Endpoint{
		Service:    service,
		Name:       "X",
		TableName:  "X",
		SchemaName: "dbo",
		FieldNames: []string{"x", "a", "d"},
		Fields:     map[string]endpoint.Field{},
	}

	x.Fields["x"] = endpoint.Field{Endpoint: x, Name: "x", FieldType: objectType.STRING, Nullable: true}
	x.Fields["a"] = endpoint.Field{Endpoint: x, Name: "a", FieldType: objectType.INTEGER, Nullable: true}
	x.Fields["d"] = endpoint.Field{Endpoint: x, Name: "d", FieldType: objectType.DATETIME, Nullable: true}

	xy := &endpoint.Endpoint{
		Service:    service,
		Name:       "XY",
		TableName:  "XY",
		SchemaName: "dbo",
		FieldNames: []string{"x", "y", "a", "b"},
		Fields:     map[string]endpoint.Field{},
	}

	xy.Fields["x"] = endpoint.Field{Endpoint: xy, Name: "x", FieldType: objectType.STRING, Nullable: true}
	xy.Fields["y"] = endpoint.Field{Endpoint: xy, Name: "y", FieldType: objectType.STRING, Nullable: true}
	xy.Fields["a"] = endpoint.Field{Endpoint: xy, Name: "a", FieldType: objectType.INTEGER, Nullable: true}
	xy.Fields["b"] = endpoint.Field{Endpoint: xy, Name: "b", FieldType: objectType.INTEGER, Nullable: true}

	yz := &endpoint.Endpoint{
		Service:    service,
		Name:       "YZ",
		TableName:  "YZ",
		SchemaName: "dbo",
		FieldNames: []string{"y", "z", "b", "c"},
		Fields:     map[string]endpoint.Field{},
	}

	yz.Fields["y"] = endpoint.Field{Endpoint: yz, Name: "y", FieldType: objectType.STRING, Nullable: true}
	yz.Fields["z"] = endpoint.Field{Endpoint: yz, Name: "z", FieldType: objectType.STRING, Nullable: true}
	yz.Fields["b"] = endpoint.Field{Endpoint: yz, Name: "b", FieldType: objectType.INTEGER, Nullable: true}
	yz.Fields["c"] = endpoint.Field{Endpoint: yz, Name: "c", FieldType: objectType.INTEGER, Nullable: true}

	service.Endpoints["X"] = x
	service.Endpoints["XY"] = xy
	service.Endpoints["YZ"] = yz

	return New(input, service.Endpoints["X"])
}
