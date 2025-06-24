package transpiler

import (
	"testing"
)

func TestBuiltinFunctions(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"x: len(@) > 5;", "SELECT X.[x] FROM dbo.X WHERE (LEN(X.[x]) > 5)"},                                         // Len function
		{"x: like(@,'%hello%');", "SELECT X.[x] FROM dbo.X WHERE (X.[x] LIKE '%hello%')"},                            // Like function
		{"d: > datetime('2025/04/03');", "SELECT X.[d] FROM dbo.X WHERE (X.[d] > CONVERT(date, '2025/04/03', 127))"}, // Datetime function
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
