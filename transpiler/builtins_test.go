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
		{"AS('uniquedate', distinct(@(DateTimeN))):",
			"SELECT (DISTINCT Types.[DateTimeN]) AS [uniquedate] FROM dbo.Types"}, // Distinct function
		{"AS('ispositive', iif(@(Int) > 0, 1, 0)):",
			"SELECT (IIF((Types.[Int] > 0), 1, 0)) AS [ispositive] FROM dbo.Types"}, // Iif function 
		{"AS('previousdate', lag(@(DateTimeN), NULL, @(DateTimeN))):",
			"SELECT (LAG(Types.[DateTimeN]) OVER (PARTITION BY NULL ORDER BY Types.[DateTimeN])) AS [previousdate] FROM dbo.Types"}, // Lag function
		{"AS('nextdatebyint', lead(@(DateTimeN), @(Int), @(DateTimeN))):",
			"SELECT (LEAD(Types.[DateTimeN]) OVER (PARTITION BY Types.[Int] ORDER BY Types.[DateTimeN])) AS [nextdatebyint] FROM dbo.Types"}, // Lead function
		{"AS('deci', cast(@('Int') , 'float' )):",
			"SELECT (CAST( Types.[Int] AS float )) AS [deci] FROM dbo.Types"}, // Cast function
		{"AS('year', convert('year', @('DateTimeN'))):",
			"SELECT (CONVERT(year, Types.[DateTimeN])) AS [year] FROM dbo.Types"}, // Convert function
		{"DateTimeN: > datetime('2025/04/03');",
			"SELECT Types.[DateTimeN] FROM dbo.Types WHERE (Types.[DateTimeN] > CONVERT(date, '2025/04/03', 127))"}, // Datetime function
		{"AS('utc', timezone(@('DateTimeN'), 'UTC')):",
			"SELECT (Types.[DateTimeN] AT TIME ZONE 'UTC') AS [utc] FROM dbo.Types"}, // Timezone function
		{"AS('year', datepart('year', @('DateTimeN'))):",
			"SELECT (DATEPART(year, Types.[DateTimeN])) AS [year] FROM dbo.Types"}, // Datepart function
		{"AS('tomorrow', dateadd('day', 1, @('DateTimeN'))):",
			"SELECT (DATEADD(day, 1, Types.[DateTimeN])) AS [tomorrow] FROM dbo.Types"}, // Dateadd function
		{"AS('daysago', datediff('day', @('DateTimeN'), date('2025-04-03'))):",
			"SELECT (DATEDIFF(day, Types.[DateTimeN], CONVERT(date, '2025-04-03', 23))) AS [daysago] FROM dbo.Types"}, // Datediff function
		{"StrN: like(@,'%hello%');",
			"SELECT Types.[StrN] FROM dbo.Types WHERE (Types.[StrN] LIKE '%hello%')"}, // Like function
		{"StrN: unlike(@, '%hello%');",
			"SELECT Types.[StrN] FROM dbo.Types WHERE (Types.[StrN] NOT LIKE '%hello%' OR Types.[StrN] IS NULL)"}, // Like inversion
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
