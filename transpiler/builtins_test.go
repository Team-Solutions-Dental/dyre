package transpiler

import (
	"testing"
)

func TestBuiltinFunctions(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"StrN: len(@) > 5;",
			"SELECT Types.[StrN] FROM dbo.Types WHERE (LEN(Types.[StrN]) > 5)"}, // Len function
		{"StrN: like(@,'%hello%');",
			"SELECT Types.[StrN] FROM dbo.Types WHERE (Types.[StrN] LIKE '%hello%')"}, // Like function
		{"DateTimeN: > datetime('2025/04/03');",
			"SELECT Types.[DateTimeN] FROM dbo.Types WHERE (Types.[DateTimeN] > CONVERT(date, '2025/04/03', 127))"}, // Datetime function
		{"AS('year', datepart('year', @('DateTimeN'))):",
			"SELECT ( DATEPART(year, Types.[DateTimeN]) ) AS year FROM dbo.Types"}, // datepart function
		{"AS('year', convert('year', @('DateTimeN'))):",
			"SELECT ( CONVERT(year, Types.[DateTimeN]) ) AS year FROM dbo.Types"}, // convert function
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
