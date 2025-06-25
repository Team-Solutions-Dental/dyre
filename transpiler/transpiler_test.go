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
		{"Int:", "SELECT Types.[Int] FROM dbo.Types"},                                                                                      // Basic Request
		{"Int:Str:Bool:", "SELECT Types.[Int], Types.[Str], Types.[Bool] FROM dbo.Types"},                                                  // Chain Column
		{"Int:;Str:;Int:", "SELECT Types.[Str], Types.[Int] FROM dbo.Types"},                                                               // Reorder
		{"Str: @ == 'Hello'", "SELECT Types.[Str] FROM dbo.Types WHERE (Types.[Str] = 'Hello')"},                                           // @ reference call
		{"Bool: @ == FALSE", "SELECT Types.[Bool] FROM dbo.Types WHERE (Types.[Bool] = 0)"},                                                // Boolean Comparison
		{"Int: Int: > 5", "SELECT Types.[Int] FROM dbo.Types WHERE (Types.[Int] > 5)"},                                                     // Integer Comparison
		{"Int: > 5 OR < 10", "SELECT Types.[Int] FROM dbo.Types WHERE ((Types.[Int] > 5) OR (Types.[Int] < 10))"},                          // OR Statement
		{"Date: @ == date('01/02/2023')", "SELECT Types.[Date] FROM dbo.Types WHERE (Types.[Date] = CONVERT(date, '01/02/2023', 23))"},     // Function Call
		{"Int2:", "SELECT Types.[Int2] FROM dbo.Types"},                                                                                    // AlphaNumeric Column
		{"@('Str') != NULL; Bool: == false;", "SELECT Types.[Bool] FROM dbo.Types WHERE (Types.[Str] IS NOT NULL) AND (Types.[Bool] = 0)"}, // @() reference function
	}

	for _, tt := range tests {
		ir, err := testNewTypes(tt.input)
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
		{"Int:", 100, "SELECT TOP 100 Types.[Int] FROM dbo.Types"},
		{"Int:", -1, "SELECT Types.[Int] FROM dbo.Types"},
	}

	for _, tt := range tests {
		evalualted_ir, err := testNewTypes(tt.input)
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

func TestComparisonExpressions(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Int: > 5", "SELECT Types.[Int] FROM dbo.Types WHERE (Types.[Int] > 5)"},                                 // GT >
		{"Int: < 5", "SELECT Types.[Int] FROM dbo.Types WHERE (Types.[Int] < 5)"},                                 // LT <
		{"Int: >= 5", "SELECT Types.[Int] FROM dbo.Types WHERE (Types.[Int] >= 5)"},                               // GTE >=
		{"Int: <= 5", "SELECT Types.[Int] FROM dbo.Types WHERE (Types.[Int] <= 5)"},                               // LTE <=
		{"Int: == 5", "SELECT Types.[Int] FROM dbo.Types WHERE (Types.[Int] = 5)"},                                // EQ ==
		{"Int: != 5", "SELECT Types.[Int] FROM dbo.Types WHERE (Types.[Int] != 5)"},                               // NOT_EQ !=
		{"Int: > 6 AND < 9", "SELECT Types.[Int] FROM dbo.Types WHERE ((Types.[Int] > 6) AND (Types.[Int] < 9))"}, // AND
		{"Int: > 6 OR < 9", "SELECT Types.[Int] FROM dbo.Types WHERE ((Types.[Int] > 6) OR (Types.[Int] < 9))"},   // OR
	}

	for _, tt := range tests {
		ir, err := testNewTypes(tt.input)
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

func TestNullExpressions(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"StrN: == NULL", "SELECT Types.[StrN] FROM dbo.Types WHERE (Types.[StrN] IS NULL)"},     // EQ == NULL
		{"StrN: != NULL", "SELECT Types.[StrN] FROM dbo.Types WHERE (Types.[StrN] IS NOT NULL)"}, // NOT_EQ != NULL
	}

	for _, tt := range tests {
		ir, err := testNewTypes(tt.input)
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

// func TestKeywords(t *testing.T) {
// 	tests := []struct {
// 		input    string
// 		expected string
// 	}{
// 		{"AS('GROUP', @('group') ):Str:", "SELECT ( 1000 ) AS GROUP, Types.[Str] FROM dbo.Types"}, // GROUP Keyword
// 	}
//
// 	for _, tt := range tests {
// 		// Defined in transpiler_test.go
// 		ir, _ := testNewTypes(tt.input)
// 		sql_statement, err := ir.EvaluateQuery()
//
// 		if err != nil {
// 			t.Errorf("Query test error. [%s] %s\n", tt.input, err.Error())
// 		}
//
// 		if sql_statement != tt.expected {
// 			t.Errorf("Query failed. [%s]\n%s \n%s\n ", tt.input, sql_statement, tt.expected)
// 		}
// 	}
// }

func testNewTypes(input string) (*PrimaryIR, error) {
	var service *endpoint.Service = &endpoint.Service{Endpoints: map[string]*endpoint.Endpoint{}}
	service.EndpointNames = []string{"Types"}

	t := &endpoint.Endpoint{
		Name:       "Types",
		TableName:  "Types",
		SchemaName: "dbo",
		FieldNames: []string{

			"Int", "IntN", "Int2",
			"Float", "FloatN",
			"Str", "StrN",
			"Bool", "BoolN",
			"Date", "DateN",
			"DateTime", "DateTimeN",
			"Group",
		},
		Fields: map[string]endpoint.Field{},
	}

	t.Fields["Int"] = endpoint.Field{Endpoint: t, Name: "Int", FieldType: objectType.INTEGER, Nullable: false}
	t.Fields["IntN"] = endpoint.Field{Endpoint: t, Name: "IntN", FieldType: objectType.INTEGER, Nullable: true}
	t.Fields["Int2"] = endpoint.Field{Endpoint: t, Name: "Int2", FieldType: objectType.INTEGER, Nullable: true}
	t.Fields["Float"] = endpoint.Field{Endpoint: t, Name: "Float", FieldType: objectType.FLOAT, Nullable: false}
	t.Fields["FloatN"] = endpoint.Field{Endpoint: t, Name: "FloatN", FieldType: objectType.FLOAT, Nullable: true}
	t.Fields["Str"] = endpoint.Field{Endpoint: t, Name: "Str", FieldType: objectType.STRING, Nullable: false}
	t.Fields["StrN"] = endpoint.Field{Endpoint: t, Name: "StrN", FieldType: objectType.STRING, Nullable: true}
	t.Fields["Group"] = endpoint.Field{Endpoint: t, Name: "Group", FieldType: objectType.STRING, Nullable: true}
	t.Fields["Bool"] = endpoint.Field{Endpoint: t, Name: "Bool", FieldType: objectType.BOOLEAN, Nullable: false}
	t.Fields["BoolN"] = endpoint.Field{Endpoint: t, Name: "BoolN", FieldType: objectType.BOOLEAN, Nullable: true}
	t.Fields["Date"] = endpoint.Field{Endpoint: t, Name: "Date", FieldType: objectType.DATE, Nullable: false}
	t.Fields["DateN"] = endpoint.Field{Endpoint: t, Name: "DateN", FieldType: objectType.DATE, Nullable: true}
	t.Fields["DateTime"] = endpoint.Field{Endpoint: t, Name: "DateTime", FieldType: objectType.DATETIME, Nullable: false}
	t.Fields["DateTimeN"] = endpoint.Field{Endpoint: t, Name: "DateTimeN", FieldType: objectType.DATETIME, Nullable: true}

	service.Endpoints["Types"] = t

	return New(input, service.Endpoints["Types"])
}

func testNewXYZ(input string) (*PrimaryIR, error) {
	var service *endpoint.Service = &endpoint.Service{Endpoints: map[string]*endpoint.Endpoint{}}
	service.EndpointNames = []string{"XN", "XYN", "YZN"}

	x := &endpoint.Endpoint{
		Service:    service,
		Name:       "XN",
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
		Name:       "XYN",
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
		Name:       "YZN",
		TableName:  "YZ",
		SchemaName: "dbo",
		FieldNames: []string{"y", "z", "b", "c"},
		Fields:     map[string]endpoint.Field{},
	}

	yz.Fields["y"] = endpoint.Field{Endpoint: yz, Name: "y", FieldType: objectType.STRING, Nullable: true}
	yz.Fields["z"] = endpoint.Field{Endpoint: yz, Name: "z", FieldType: objectType.STRING, Nullable: true}
	yz.Fields["b"] = endpoint.Field{Endpoint: yz, Name: "b", FieldType: objectType.INTEGER, Nullable: true}
	yz.Fields["c"] = endpoint.Field{Endpoint: yz, Name: "c", FieldType: objectType.INTEGER, Nullable: true}

	service.Endpoints["XN"] = x
	service.Endpoints["XYN"] = xy
	service.Endpoints["YZN"] = yz

	return New(input, service.Endpoints["XN"])
}
