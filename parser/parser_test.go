package parser

import (
	"fmt"
	"testing"

	"github.com/vamuscari/dyre/ast"
	"github.com/vamuscari/dyre/lexer"
)

func TestChainColumnStatement(t *testing.T) {
	input := "foo:bar:baz:"
	l := lexer.New(input)
	p := New(l)
	query := p.ParseQuery()
	checkParserErrors(t, p)

	if len(query.Statements) != 3 {
		t.Errorf("query.Statements does not contain %d statements, got=%d\n",
			3, len(query.Statements))
	}

	for i := range query.Statements {
		clmn, ok := query.Statements[i].(*ast.ColumnStatement)
		if !ok {
			t.Fatalf("query.Statements[%d] is not ast.ColumnStatement. got=%T", i, query.Statements[i])
		}
		if clmn.Expressions != nil {
			t.Fatalf("query.Statements[%d].expression is not nil", i)
		}
	}
}

func TestChainColumnStatementWithExpression(t *testing.T) {
	input := "foo: > 5; < 10;bar: > 0; < 3;"
	l := lexer.New(input)
	p := New(l)
	query := p.ParseQuery()
	checkParserErrors(t, p)

	if len(query.Statements) != 2 {
		t.Errorf("query.Statements does not contain %d statements, got=%d\n",
			2, len(query.Statements))
		for i, s := range query.Statements {
			t.Errorf("Statement[%d]: %s", i, s.String())
		}
	}

	for i := range query.Statements {
		clmn, ok := query.Statements[i].(*ast.ColumnStatement)
		if !ok {
			fmt.Printf("\n %s", query.Statements[i].TokenLiteral())
			t.Fatalf("query.Statements[%d] is not ast.ColumnStatement. got=%T", i, query.Statements[i])
		}

		if len(clmn.Expressions.Statements) != 2 {
			for j, es := range clmn.Expressions.Statements {
				t.Errorf("Query %d. Expression %d. %s", i, j, es.String())
			}
			t.Fatalf("query.Statements[%d].expression.statements is not 2, got=%d", i, len(clmn.Expressions.Statements))
		}

	}
}

func TestColumnStatement(t *testing.T) {
	input := "foo: 5; false; func(@);"
	l := lexer.New(input)
	p := New(l)
	query := p.ParseQuery()
	checkParserErrors(t, p)

	if len(query.Statements) != 1 {
		t.Errorf("query.Statements does not contain %d statements, got=%d\n",
			1, len(query.Statements))
	}

	clmn, ok := query.Statements[0].(*ast.ColumnStatement)
	if !ok {
		t.Fatalf("query.Statements[0] is not ast.ColumnStatement. got=%T", query.Statements[0])
	}

	if len(clmn.Expressions.Statements) != 3 {
		t.Errorf("clmn.Expressions.Statements does not contain %d statements, got=%d\n",
			2, len(query.Statements))
	}

	stmnt0, ok := clmn.Expressions.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("Statements[0] is not ast.ExpressionStatement. got=%T", clmn.Expressions.Statements[0])
	}

	if !testIntegerLiteral(t, stmnt0.Expression, 5) {
		return
	}

	stmnt1, ok := clmn.Expressions.Statements[1].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("Statements[1] is not ast.ExpressionStatement. got=%T", clmn.Expressions.Statements[1])
	}

	if !testBooleanLiteral(t, stmnt1.Expression, false) {
		return
	}
}

func TestIdentifierExpression(t *testing.T) {
	input := "foobar"

	l := lexer.New(input)
	p := New(l)
	query := p.ParseQuery()
	checkParserErrors(t, p)

	if len(query.Statements) != 1 {
		t.Fatalf("query has not enough statemnets. got=%d", len(query.Statements))
	}
	stmt, ok := query.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("query.Statements[0] is not ast.ExpressionStatement. got=%T", query.Statements[0])
	}
	ident, ok := stmt.Expression.(*ast.Identifier)
	if !ok {
		t.Fatalf("exp not *ast.Identifier. got=%T", stmt.Expression)
	}
	if ident.Value != "foobar" {
		t.Errorf("ident.Value not %s. got=%s", "foobar", ident.Value)
	}
	if ident.TokenLiteral() != "foobar" {
		t.Errorf("ident.TokenLiteral not %s. got=%s", "foobar", ident.TokenLiteral())
	}
}

func TestIntegerLiteralExpression(t *testing.T) {
	input := "5"

	l := lexer.New(input)
	p := New(l)
	query := p.ParseQuery()
	checkParserErrors(t, p)

	if len(query.Statements) != 1 {
		t.Fatalf("query has not enough statements. got=%d",
			len(query.Statements))
	}

	stmt, ok := query.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("query.Statements[0] is not ast.ExpressionStatement. got=%T",
			query.Statements[0])
	}

	literal, ok := stmt.Expression.(*ast.IntegerLiteral)
	if !ok {
		t.Fatalf("exp not *ast.IntegerLiteral. got=%T", stmt.Expression)
	}
	if literal.Value != 5 {
		t.Errorf("literal.Value not %d. got=%d", 5, literal.Value)
	}
	if literal.TokenLiteral() != "5" {
		t.Errorf("literal.TokenLiteral not %s. got=%s", "5", literal.TokenLiteral())
	}
}

func TestBooleanExpression(t *testing.T) {
	input := "true"

	l := lexer.New(input)
	p := New(l)
	query := p.ParseQuery()
	checkParserErrors(t, p)

	if len(query.Statements) != 1 {
		t.Fatalf("query has not enough statements. got=%d", len(query.Statements))
	}

	stmt, ok := query.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("query.Statements[0] is not ast.ExpressionStatement. got=%T",
			query.Statements[0])
	}

	literal, ok := stmt.Expression.(*ast.Boolean)
	if !ok {
		t.Fatalf("exp not *ast.BooleanLiteral. got=%T", stmt.Expression)
	}
	if literal.Value != true {
		t.Errorf("literal.Value not '%t'. got=%t", true, literal.Value)
	}
	if literal.TokenLiteral() != "true" {
		t.Errorf("literal.TokenLiteral not %s. got=%s", "true", literal.TokenLiteral())
	}

}

func TestParsingPrefixExpression(t *testing.T) {
	prefixTest := []struct {
		input    string
		operator string
		value    interface{}
	}{
		{"!5", "!", 5},
		{"-15", "-", 15},
		{"!true", "!", true},
		{"!false", "!", false},
	}

	for _, tt := range prefixTest {
		l := lexer.New(tt.input)
		p := New(l)
		query := p.ParseQuery()
		checkParserErrors(t, p)

		if len(query.Statements) != 1 {
			t.Fatalf("query.Statements does not contain %d statements. got =%d",
				1, len(query.Statements))
		}

		stmt, ok := query.Statements[0].(*ast.ExpressionStatement)
		if !ok {
			t.Fatalf("query.Statements[0] is not ast.ExpressionStatement, got=%T",
				query.Statements[0])
		}

		if !testPrefixExpression(t, stmt.Expression, tt.operator, tt.value) {
			return
		}
	}
}

func TestParsingInfixExpressions(t *testing.T) {
	infixTests := []struct {
		input      string
		leftValue  interface{}
		operator   string
		rightValue interface{}
	}{
		{"5 + 5", 5, "+", 5},
		{"5 - 5", 5, "-", 5},
		{"5 * 5", 5, "*", 5},
		{"5 / 5", 5, "/", 5},
		{"5 > 5", 5, ">", 5},
		{"5 < 5", 5, "<", 5},
		{"5 == 5", 5, "==", 5},
		{"5 != 5", 5, "!=", 5},
		{"true == true", true, "==", true},
		{"true != false", true, "!=", false},
		{"false == false", false, "==", false},
		{"5 AND 5", 5, "AND", 5},
	}

	for _, tt := range infixTests {
		l := lexer.New(tt.input)
		p := New(l)
		query := p.ParseQuery()
		checkParserErrors(t, p)

		if len(query.Statements) != 1 {
			t.Fatalf("query.Statements does not contain %d statements. got =%d",
				1, len(query.Statements))
		}

		stmt, ok := query.Statements[0].(*ast.ExpressionStatement)
		if !ok {
			t.Fatalf("query.Statements[0] is not ast.ExpressionStatement, got=%T",
				query.Statements[0])
		}

		if !testInfixExpresssion(t, stmt.Expression, tt.leftValue, tt.operator, tt.rightValue) {
			return
		}

	}
}

func testIntegerLiteral(t *testing.T, il ast.Expression, value int64) bool {
	integ, ok := il.(*ast.IntegerLiteral)
	if !ok {
		t.Errorf("il no *ast.IntegerLiteral. got=%T", il)
		return false
	}

	if integ.Value != value {
		t.Errorf("integ.Value not %d. got=%d",
			value, integ.Value)
		return false
	}

	if integ.TokenLiteral() != fmt.Sprintf("%d", value) {
		t.Errorf("integ.TokenLiteral not %d. got=%s",
			value, integ.TokenLiteral())
		return false
	}

	return true
}

func checkParserErrors(t *testing.T, p *Parser) {
	errors := p.Errors()
	if len(errors) == 0 {
		return
	}

	t.Errorf("parser has %d errors", len(errors))
	for _, msg := range errors {
		t.Errorf("parser error: %q", msg)
	}
	t.FailNow()
}

func TestOperatorPrecedenceParsing(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			"-a * b",
			"((-a) * b)",
		},
		{
			"!-a",
			"(!(-a))",
		},
		{
			"a + b + c",
			"((a + b) + c)",
		},
		{
			"a + b - c",
			"((a + b) - c)",
		},
		{
			"a * b * c",
			"((a * b) * c)",
		},
		{
			"a * b / c",
			"((a * b) / c)",
		},
		{
			"a + b / c",
			"(a + (b / c))",
		},
		{
			"a + b * c + d / e - f",
			"(((a + (b * c)) + (d / e)) - f)",
		},
		{
			"3 + 4; -5 * 5",
			"(3 + 4)((-5) * 5)",
		},
		{
			"5 > 4 != 3 > 4",
			"((5 > 4) != (3 > 4))",
		},
		{
			"5 < 4 != 3 > 4",
			"((5 < 4) != (3 > 4))",
		},
		{
			"3 + 4 * 5 == 3 * 1 + 4 * 5",
			"((3 + (4 * 5)) == ((3 * 1) + (4 * 5)))",
		},
		{
			"true",
			"true",
		},
		{
			"false",
			"false",
		},
		{
			"3 > 5 == false",
			"((3 > 5) == false)",
		},
		{
			"3 < 5 == true",
			"((3 < 5) == true)",
		},
		{
			"1 + (2 + 3) + 4",
			"((1 + (2 + 3)) + 4)",
		},
		{
			"(5 + 5) * 2",
			"((5 + 5) * 2)",
		},
		{
			"2 / (5 + 5)",
			"(2 / (5 + 5))",
		},
		{
			"-(5 + 5)",
			"(-(5 + 5))",
		},
		{
			"!(true == true)",
			"(!(true == true))",
		},
		{
			"a + add(b * c) + d",
			"((a + add((b * c))) + d)",
		},
		{
			"add(a, b, 1, 2 * 3, 4 + 5, add(6, 7 * 8))",
			"add(a, b, 1, (2 * 3), (4 + 5), add(6, (7 * 8)))",
		},
		{
			"add(a + b + c * d / f + g)",
			"add((((a + b) + ((c * d) / f)) + g))",
		},
		{
			"true == true && false == false",
			"((true == true) && (false == false))",
		},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		query := p.ParseQuery()
		checkParserErrors(t, p)

		actual := query.String()
		if actual != tt.expected {
			t.Errorf("expected=%q, got=%q", tt.expected, actual)
		}
	}
}

func testIdentifier(t *testing.T, exp ast.Expression, value string) bool {
	ident, ok := exp.(*ast.Identifier)
	if !ok {
		t.Errorf("exp not *ast.Identifier. got=%T", exp)
		return false
	}

	if ident.Value != value {
		t.Errorf("ident.Value not %s. got=%s", value, ident.Value)
		return false
	}

	if ident.TokenLiteral() != value {
		t.Errorf("ident.TokenLiteral not %s. got=%s", value, ident.TokenLiteral())
		return false
	}

	return true
}

func testLiteralExpression(t *testing.T, exp ast.Expression, expected interface{}) bool {
	switch v := expected.(type) {
	case int:
		return testIntegerLiteral(t, exp, int64(v))
	case int64:
		return testIntegerLiteral(t, exp, v)
	case string:
		return testIdentifier(t, exp, v)
	case bool:
		return testBooleanLiteral(t, exp, v)
	}
	t.Errorf("type of exp not handled. got=%T", exp)
	return false
}

func testInfixExpresssion(t *testing.T,
	exp ast.Expression,
	left interface{},
	operator string,
	right interface{}) bool {
	opExp, ok := exp.(*ast.InfixExpression)
	if !ok {
		t.Errorf("exp is not ast.InfixExpression. got=%T(%s)", exp, exp)
		return false
	}

	if !testLiteralExpression(t, opExp.Left, left) {
		return false
	}

	if opExp.Operator != operator {
		t.Errorf("exp.Operator is not '%s'. got=%q", operator, opExp.Operator)
		return false
	}

	if !testLiteralExpression(t, opExp.Right, right) {
		return false
	}

	return true
}

func testPrefixExpression(t *testing.T,
	exp ast.Expression,
	operator string,
	right interface{},
) bool {
	opExp, ok := exp.(*ast.PrefixExpression)
	if !ok {
		t.Errorf("exp is not ast.PrefixExpression. got=%T(%s)", exp, exp)
		return false
	}

	if opExp.Operator != operator {
		t.Errorf("exp.Operator is not '%s'. got=%q", operator, opExp.Operator)
		return false
	}

	if !testLiteralExpression(t, opExp.Right, right) {
		return false
	}

	return true
}

func testBooleanLiteral(t *testing.T, exp ast.Expression, value bool) bool {
	bo, ok := exp.(*ast.Boolean)
	if !ok {
		t.Errorf("exp not *ast.Boolean. got=%T", exp)
		return false
	}

	if bo.Value != value {
		t.Errorf("bo.Value not %t. got=%t", value, bo.Value)
		return false
	}

	if bo.TokenLiteral() != fmt.Sprintf("%t", value) {
		t.Errorf("bo.TokenLiteral not %t. got=%s", value, bo.TokenLiteral())
		return false
	}

	return true
}

func TestStringLiteralExpression(t *testing.T) {
	input := `'hello world'`

	l := lexer.New(input)
	p := New(l)
	query := p.ParseQuery()
	checkParserErrors(t, p)

	stmt := query.Statements[0].(*ast.ExpressionStatement)
	literal, ok := stmt.Expression.(*ast.StringLiteral)
	if !ok {
		t.Fatalf("exp not *ast.StringLiteral. got=%T", stmt.Expression)
	}

	if literal.Value != "hello world" {
		t.Errorf("literal.Value not %q. got=%q, ", "hello word", literal.Value)
	}

}

func TestOrderExpression(t *testing.T) {
	input := `ASC`

	l := lexer.New(input)
	p := New(l)
	query := p.ParseQuery()
	checkParserErrors(t, p)

	stmt := query.Statements[0].(*ast.ExpressionStatement)
	literal, ok := stmt.Expression.(*ast.OrderExpression)
	if !ok {
		t.Fatalf("exp not *ast.StringLiteral. got=%T", stmt.Expression)
	}

	if literal.TokenLiteral() != "ASC" {
		t.Errorf("literal not %q. got=%q, ", "hello word", literal.TokenLiteral())
	}

	if literal.Ascending != true {
		t.Errorf("literal.Ascending not %t. got=%t, ", true, literal.Ascending)
	}

}
