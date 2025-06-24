package transpiler

import (
	"testing"
)

func TestColumnFunctions(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"AS('NewName', @('string')):", "SELECT ( Test.[string] ) AS NewName FROM dbo.Test"},              // AS Function
		{"string:EXCLUDE('bool'):!=False", "SELECT Test.[string] FROM dbo.Test WHERE (Test.[bool] != 0)"}, // EXCLUDE Function
	}

	for _, tt := range tests {
		// Defined in transpiler_test.go
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
