package parser

import (
	"fmt"
	"testing"

	"github.com/vamuscari/dyre/ast"
	"github.com/vamuscari/dyre/lexer"
)

func TestColumnStatement(t *testing.T) {
	input := "foo: 5; false; func(@);"
	l := lexer.New(input)
	p := New(l)
	query := p.ParseQuery()
	checkParserErrors(t, p)

	if len(query.Statements) != 4 {
		t.Errorf("query.Statements does not contain %d statements, got=%d\n",
			4, len(query.Statements))
	}

	_, ok := query.Statements[0].(*ast.ColumnLiteral)
	if !ok {
		t.Fatalf("query.Statements[0] is not ast.ColumnLiteral. got=%T", query.Statements[0])
	}

}

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
		_, ok := query.Statements[i].(*ast.ColumnLiteral)
		if !ok {
			t.Fatalf("query.Statements[%d] is not ast.ColumnLiteral. got=%T", i, query.Statements[i])
		}
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
		t.Fatalf("exp not *ast.OrderExpression. got=%T", stmt.Expression)
	}

	if literal.TokenLiteral() != "ASC" {
		t.Errorf("literal not %q. got=%q, ", "hello word", literal.TokenLiteral())
	}

	if literal.Ascending != true {
		t.Errorf("literal.Ascending not %t. got=%t, ", true, literal.Ascending)
	}

}

func TestReferenceLiteral(t *testing.T) {
	input := `@`

	l := lexer.New(input)
	p := New(l)
	query := p.ParseQuery()
	checkParserErrors(t, p)

	stmt := query.Statements[0].(*ast.ExpressionStatement)
	ref, ok := stmt.Expression.(*ast.Reference)
	if !ok {
		t.Fatalf("exp not *ast.Reference. got=%T", stmt.Expression)
	}

	if ref.TokenLiteral() != "@" {
		t.Errorf("literal not %q. got=%q, ", "hello word", ref.TokenLiteral())
	}

	if ref.Argument != nil {
		t.Errorf("got unexpected parameter. got=%v, ", ref.Argument)
	}

	if ref.String() != input {
		t.Errorf("got unexpected parsing string . got=%s, ", ref.String())
	}

}

func TestReferenceFunction(t *testing.T) {
	input := `@('column')`

	l := lexer.New(input)
	p := New(l)
	query := p.ParseQuery()
	checkParserErrors(t, p)

	stmt := query.Statements[0].(*ast.ExpressionStatement)
	ref, ok := stmt.Expression.(*ast.Reference)
	if !ok {
		t.Fatalf("exp not *ast.Reference. got=%T", stmt.Expression)
	}

	if ref.TokenLiteral() != "@" {
		t.Errorf("literal not '%q'. got='%q', ", "hello word", ref.TokenLiteral())
	}

	if ref.Argument == nil {
		t.Errorf("Missing Argument.")
	}

	if ref.String() != input {
		t.Errorf("got unexpected parsing string . got=%s, ", ref.String())
	}

}

func TestColumnFunction(t *testing.T) {
	input := `AS('year', datepart('year', @('CreateDate'))): `

	l := lexer.New(input)
	p := New(l)
	query := p.ParseQuery()
	checkParserErrors(t, p)

	fn, ok := query.Statements[0].(*ast.ColumnFunction)
	if !ok {
		t.Fatalf("exp not *ast.Reference. got=%T", query.Statements[0])
	}

	if fn.TokenLiteral() != "AS" {
		t.Errorf("literal not '%q'. got='%q', ", "AS", fn.TokenLiteral())
	}

	if fn.Arguments == nil {
		t.Errorf("Missing Arguments.")
	}

	if fn.String() != input {
		t.Errorf("got unexpected parsing output string. \nwant = '%s' \ngot  = '%s'", input, fn.String())
	}

}

func TestGroupFunction(t *testing.T) {
	input := `GROUP('Year'): `

	l := lexer.New(input)
	p := New(l)
	query := p.ParseQuery()
	checkParserErrors(t, p)

	fn, ok := query.Statements[0].(*ast.GroupFunction)
	if !ok {
		t.Fatalf("exp not *ast.Reference. got=%T", query.Statements[0])
	}

	if fn.TokenLiteral() != "GROUP" {
		t.Errorf("literal not '%q'. got='%q', ", "GROUP", fn.TokenLiteral())
	}

	if fn.Arguments == nil {
		t.Errorf("Missing Arguments.")
	}

	if fn.String() != input {
		t.Errorf("got unexpected parsing output string. \nwant = '%s' \ngot  = '%s'", input, fn.String())
	}

}
