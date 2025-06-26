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
			"SELECT (Types.[Str]) AS [NewName] FROM dbo.Types"}, // AS/Alias Function
		{"Str:AS('NewName', @('Int')):>5",
			"SELECT Types.[Str], Types.[NewName] FROM ( SELECT Types.[Str], (Types.[Int]) AS [NewName] FROM dbo.Types ) AS Types WHERE (Types.[NewName] > 5)"}, // Alias Condition
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

func TestColumnFunctionJoins(t *testing.T) {
	tests := []struct {
		input_x  string
		input_yz string
		expected string
	}{
		{"x:", "AS('x', @('z')):", "SELECT X.[x] FROM dbo.X INNER JOIN ( SELECT (YZ.[z]) AS [x] FROM dbo.YZ ) AS YZN ON X.[x] = YZN.[x]"},
		{"x:", "AS('x', @('z')): != NULL", "SELECT X.[x] FROM dbo.X INNER JOIN ( SELECT YZN.[x] FROM ( SELECT (YZ.[z]) AS [x] FROM dbo.YZ ) AS YZN WHERE (YZN.[x] IS NOT NULL) ) AS YZN ON X.[x] = YZN.[x]"},
	}

	for _, tt := range tests {
		// Defined in transpiler_test.go
		x, err := testNewXYZ(tt.input_x)
		if err != nil {
			t.Fatalf("testNewXYZ. %s\n", err.Error())
		}

		_, err = x.INNERJOIN("YZN").ON("x", "x").Query(tt.input_yz)
		if err != nil {
			t.Fatalf("INNERJOIN YZN. %s\n", err.Error())
		}

		sql_statement, err := x.EvaluateQuery()
		if err != nil {
			t.Errorf("x.EvaluateQuery() %s\n", err.Error())
		}

		if sql_statement != tt.expected {
			t.Errorf("Group failed. [%s] [%s]\n%s \n%s\n ", tt.input_x, tt.input_yz, sql_statement, tt.expected)
		}
	}
}
