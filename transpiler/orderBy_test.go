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
		{"int:", "int: ASC;", "SELECT Test.[int] FROM dbo.Test ORDER BY int ASC"},
		{"int:string:bool:", "bool: DESC", "SELECT Test.[int], Test.[string], Test.[bool] FROM dbo.Test ORDER BY bool DESC"},
		{"int:;string:;int:", "int:", "SELECT Test.[string], Test.[int] FROM dbo.Test ORDER BY int ASC"},
		{"int:;string:;int:", "int:ASC;string:DESC;", "SELECT Test.[string], Test.[int] FROM dbo.Test ORDER BY int ASC, string DESC"},
	}

	for _, tt := range tests {
		ir, err := testNew(tt.query)
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
