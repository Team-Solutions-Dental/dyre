package transpiler

import (
	"testing"
)

func TestGroupFunctions(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"GROUP('StrN'):", "SELECT Types.[StrN] FROM dbo.Types GROUP BY Types.[StrN]"},                                                                                                          // GROUP Function
		{"GROUP('year', datepart('year', @('DateTimeN'))):", "SELECT (DATEPART(year, Types.[DateTimeN])) AS [year] FROM dbo.Types GROUP BY DATEPART(year, Types.[DateTimeN])"},                  // GROUP Function
		{"GROUP('StrN'):@('Int')>5;", "SELECT Types.[StrN] FROM dbo.Types WHERE (Types.[Int] > 5) GROUP BY Types.[StrN]"},                                                                       // WHERE Evaluation
		{"GROUP('StrN'):COUNT('countedBool', @('Bool')):", "SELECT Types.[StrN], COUNT(Types.[Bool]) AS [countedBool] FROM dbo.Types GROUP BY Types.[StrN]"},                                    // COUNT Function
		{"GROUP('StrN'):SUM('sumInt', @('Int')):", "SELECT Types.[StrN], SUM(Types.[Int]) AS [sumInt] FROM dbo.Types GROUP BY Types.[StrN]"},                                                    // SUM Function
		{"GROUP('StrN'):AVG('avgInt', @('Int')):", "SELECT Types.[StrN], AVG(Types.[Int]) AS [avgInt] FROM dbo.Types GROUP BY Types.[StrN]"},                                                    // AVG Function
		{"GROUP('StrN'):MIN('minInt', @('Int')):", "SELECT Types.[StrN], MIN(Types.[Int]) AS [minInt] FROM dbo.Types GROUP BY Types.[StrN]"},                                                    // MIN Function
		{"GROUP('StrN'):MAX('maxInt', @('Int')):", "SELECT Types.[StrN], MAX(Types.[Int]) AS [maxInt] FROM dbo.Types GROUP BY Types.[StrN]"},                                                    // MAX Function
		{"GROUP('StrN'): != NULL;", "SELECT Types.[StrN] FROM dbo.Types GROUP BY Types.[StrN] HAVING (Types.[StrN] IS NOT NULL)"},                                                               // Select Field HAVING
		{"GROUP('StrN'):MAX('maxInt', (@('Int') * 10)): > 5;", "SELECT Types.[StrN], MAX((Types.[Int] * 10)) AS [maxInt] FROM dbo.Types GROUP BY Types.[StrN] HAVING ((Types.[Int] * 10) > 5)"}, // MAX Function
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

func TestGroupJoins(t *testing.T) {
	tests := []struct {
		input_x  string
		input_xy string
		expected string
	}{
		{"GROUP('b'):", "b:", "SELECT XYN.[b] FROM dbo.X INNER JOIN ( SELECT XY.[b], XY.[a] FROM dbo.XY ) AS XYN ON X.[a] = XYN.[a] GROUP BY XYN.[b]"},
		{"GROUP('b'):SUM('z',@('y')):", "b:y:", "SELECT XYN.[b], SUM(XYN.[y]) AS [z] FROM dbo.X INNER JOIN ( SELECT XY.[b], XY.[y], XY.[a] FROM dbo.XY ) AS XYN ON X.[a] = XYN.[a] GROUP BY XYN.[b]"},
	}

	for _, tt := range tests {
		// Defined in transpiler_test.go
		x, err := testNewXYZ(tt.input_x)
		if err != nil {
			t.Fatalf("testNewXYZ. %s\n", err.Error())
		}

		_, err = x.INNERJOIN("XYN").ON("a", "a").Query(tt.input_xy)
		if err != nil {
			t.Fatalf("INNERJOIN XY. %s\n", err.Error())
		}

		sql_statement, err := x.EvaluateQuery()
		if err != nil {
			t.Errorf("x.EvaluateQuery() %s\n", err.Error())
		}

		if sql_statement != tt.expected {
			t.Errorf("Group failed. [%s] [%s]\n%s \n%s\n ", tt.input_x, tt.input_xy, sql_statement, tt.expected)
		}
	}
}
