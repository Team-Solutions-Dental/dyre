package transpiler

import (
	"testing"
)

func TestOrderBy(t *testing.T) {
	tests := []struct {
		query    string
		orderBy  string
		expected string
	}{
		{"int:", "int: ASC;", "SELECT Test.[int] FROM dbo.Test ORDER BY int ASC"},
		{"int:string:bool:", "bool: DESC", "SELECT Test.[int], Test.[string], Test.[bool] FROM dbo.Test ORDER BY bool DESC"},
		{"int:;string:;int:", "int:", "SELECT Test.[string], Test.[int] FROM dbo.Test ORDER BY int ASC"},
	}

	for _, tt := range tests {
		ir, err := testNew(tt.query)
		if err != nil {
			t.Errorf("Query table error. [%s] %s\n", tt.query, err.Error())
		}

		ir.OrderBy(tt.orderBy)

		sql_statement, err := ir.EvaluateQuery()
		if err != nil {
			t.Errorf("Query evaluation error. [%s] %s\n", tt.query, err.Error())
		}

		if sql_statement != tt.expected {
			t.Errorf("Query failed. [%s]\n%s \n%s\n ", tt.query, sql_statement, tt.expected)
		}
	}
}

// func testNew(input string) (*PrimaryIR, error) {
// 	service := &endpoint.Service{
// 		Endpoints: map[string]*endpoint.Endpoint{
// 			"Test": {
// 				Name:       "Test",
// 				TableName:  "Test",
// 				SchemaName: "dbo",
// 				FieldNames: []string{"int", "string", "bool", "date"},
// 				Fields: map[string]endpoint.Field{
// 					"int":    {Name: "int", FieldType: object.INTEGER_OBJ, Nullable: true},
// 					"string": {Name: "string", FieldType: object.STRING_OBJ, Nullable: true},
// 					"bool":   {Name: "bool", FieldType: object.BOOLEAN_OBJ, Nullable: true},
// 					"date":   {Name: "date", FieldType: object.DATE_OBJ, Nullable: true},
// 				},
// 			},
// 		},
// 	}
// 	return New(input, service.Endpoints["Test"])
// }
