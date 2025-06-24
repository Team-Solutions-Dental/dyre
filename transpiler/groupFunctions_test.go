package transpiler

import (
	"testing"
)

func TestGroupFunctions(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"GROUP('Str'):", "SELECT Types.[Str] FROM dbo.Types GROUP BY Types.[Str]"},                                                                       // GROUP Function
		{"GROUP('Str'):@('Int')>5;", "SELECT Types.[Str] FROM dbo.Types WHERE (Types.[Int] > 5) GROUP BY Types.[Str]"},                                    // WHERE Evaluation
		{"GROUP('Str'):COUNT('countedBool', @('Bool')):", "SELECT Types.[Str], COUNT( Types.[Bool] ) AS countedBool FROM dbo.Types GROUP BY Types.[Str]"}, // COUNT Function
		{"GROUP('Str'):SUM('sumInt', @('Int')):", "SELECT Types.[Str], SUM( Types.[Int] ) AS sumInt FROM dbo.Types GROUP BY Types.[Str]"},                 // SUM Function
		{"GROUP('Str'):AVG('avgInt', @('Int')):", "SELECT Types.[Str], AVG( Types.[Int] ) AS avgInt FROM dbo.Types GROUP BY Types.[Str]"},                 // AVG Function
		{"GROUP('Str'):MIN('minInt', @('Int')):", "SELECT Types.[Str], MIN( Types.[Int] ) AS minInt FROM dbo.Types GROUP BY Types.[Str]"},                 // MIN Function
		{"GROUP('Str'):MAX('maxInt', @('Int')):", "SELECT Types.[Str], MAX( Types.[Int] ) AS maxInt FROM dbo.Types GROUP BY Types.[Str]"},                 // MAX Function
		{"GROUP('Str'): != NULL;", "SELECT Types.[Str] FROM dbo.Types GROUP BY Types.[Str] HAVING (Types.[Str] IS NOT NULL)"},                             // Select Field HAVING
	}

	for _, tt := range tests {
		// Defined in transpiler_test.go
		ir, _ := testNewTypes(tt.input)
		sql_statement, err := ir.EvaluateQuery()

		if err != nil {
			t.Errorf("Group test error. [%s] %s\n", tt.input, err.Error())
		}

		if sql_statement != tt.expected {
			t.Errorf("Group failed. [%s]\n%s \n%s\n ", tt.input, sql_statement, tt.expected)
		}
	}
}

func TestGroupError(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"GROUP('Str'):Bool:", "ERROR: Column 'Bool' cannot be called on Grouped Table 'Types'"},
		{"Bool:GROUP('Str'):", "ERROR: Group Function 'GROUP' cannot be called on Non-Grouped Table 'Types'"},
	}

	for _, tt := range tests {
		// Defined in transpiler_test.go
		ir, _ := testNewTypes(tt.input)
		_, err := ir.EvaluateQuery()

		if err == nil {
			t.Errorf("Group error test. %s\n%s\n", "Missing error", tt.input)
		}

		if err.Error() != tt.expected {
			t.Errorf("Group error failed. [%s]\n%s \n%s\n ", tt.input, err.Error(), tt.expected)
		}
	}
}
