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
		{"Int:", "Int: ASC;", "SELECT Types.[Int] FROM dbo.Types ORDER BY Int ASC"},
		{"Int:Str:Bool:", "Bool: DESC", "SELECT Types.[Int], Types.[Str], Types.[Bool] FROM dbo.Types ORDER BY Bool DESC"},
		{"Int:;Str:;Int:", "Int:", "SELECT Types.[Str], Types.[Int] FROM dbo.Types ORDER BY Int ASC"},
		{"Int:;Str:;Int:", "Int:ASC;Str:DESC;", "SELECT Types.[Str], Types.[Int] FROM dbo.Types ORDER BY Int ASC, Str DESC"},
	}

	for _, tt := range tests {
		ir, err := testNewTypes(tt.query)
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
