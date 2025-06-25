package transpiler

import (
	"testing"
)

func TestColumnFunctions(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"AS('NewName', @('Str')):",
			"SELECT ( Types.[Str] ) AS NewName FROM dbo.Types"}, // AS/Alias Function
		{"Str:AS('NewName', @('Int')):>5",
			"SELECT Types.[Str], Types.[NewName] FROM ( SELECT Types.[Str], ( Types.[Int] ) AS NewName FROM dbo.Types ) AS Types WHERE (Types.[NewName] > 5)"}, // Alias Condition
		{"Str:EXCLUDE('Bool'):!=False",
			"SELECT Types.[Str] FROM dbo.Types WHERE (Types.[Bool] != 0)"}, // EXCLUDE Function
	}

	for _, tt := range tests {
		// Defined in transpiler_test.go
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
