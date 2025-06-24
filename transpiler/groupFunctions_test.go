package transpiler

import (
	"testing"
)

func TestGroupFunctions(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"GROUP('string'):", "SELECT Test.[string] FROM dbo.Test GROUP BY Test.[string]"},                                                                      // GROUP Function
		{"GROUP('string'):@('int')>5;", "SELECT Test.[string] FROM dbo.Test WHERE (Test.[int] > 5) GROUP BY Test.[string]"},                                    // WHERE Evaluation
		{"GROUP('string'):COUNT('countedBool', @('bool')):", "SELECT Test.[string], COUNT( Test.[bool] ) AS countedBool FROM dbo.Test GROUP BY Test.[string]"}, // COUNT Function
		{"GROUP('string'):SUM('sumINT', @('int')):", "SELECT Test.[string], SUM( Test.[int] ) AS sumINT FROM dbo.Test GROUP BY Test.[string]"},                 // SUM Function
		{"GROUP('string'):AVG('avgINT', @('int')):", "SELECT Test.[string], AVG( Test.[int] ) AS avgINT FROM dbo.Test GROUP BY Test.[string]"},                 // AVG Function
		{"GROUP('string'):MIN('minINT', @('int')):", "SELECT Test.[string], MIN( Test.[int] ) AS minINT FROM dbo.Test GROUP BY Test.[string]"},                 // MIN Function
		{"GROUP('string'):MAX('maxINT', @('int')):", "SELECT Test.[string], MAX( Test.[int] ) AS maxINT FROM dbo.Test GROUP BY Test.[string]"},                 // MAX Function
		{"GROUP('string'): != NULL;", "SELECT Test.[string] FROM dbo.Test GROUP BY Test.[string] HAVING (Test.[string] IS NOT NULL)"},                          // Select Field HAVING
	}

	for _, tt := range tests {
		// Defined in transpiler_test.go
		ir, _ := testNew(tt.input)
		sql_statement, err := ir.EvaluateQuery()

		if err != nil {
			t.Errorf("Query test error. [%s] %s\n", tt.input, err.Error())
		}

		if sql_statement != tt.expected {
			t.Errorf("Query failed. [%s]\n%s \n%s\n ", tt.input, sql_statement, tt.expected)
		}
	}
}

func TestGroupError(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"GROUP('string'):bool:", "ERROR: Column 'bool' cannot be called on Grouped Table 'Test'"},
		{"bool:GROUP('string'):", "ERROR: Group 'GROUP' cannot be called on Non-Grouped Table 'Test'"},
	}

	for _, tt := range tests {
		// Defined in transpiler_test.go
		ir, _ := testNew(tt.input)
		sql_statement, err := ir.EvaluateQuery()

		if err == nil {
			t.Errorf("Group error test. %s\n%s\n", "Missing error", tt.input)
		}

		if err.Error() != tt.expected {
			t.Errorf("Group error failed. [%s]\n%s \n%s\n ", tt.input, sql_statement, tt.expected)
		}
	}
}
