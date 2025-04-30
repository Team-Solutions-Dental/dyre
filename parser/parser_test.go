package parser

import (
	"fmt"
	"testing"

	"github.com/vamuscari/dyre/ast"
	"github.com/vamuscari/dyre/lexer"
)

func TestColumnStatements(t *testing.T) {
	tests := []struct {
		input              string
		expectedIdentifier string
		expectedValue      interface{}
	}{
		{"foo: x;", "foo", "x"},
		{"bar: true;", "bar", true},
		{"bar: true;", "bar", true},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		query := p.ParseQuery()
		checkParserErrors(t, p)

		if len(query.Statements) != 1 {
			t.Fatalf("query.Statements does not contain 1 statemnets. got=%d",
				len(query.Statements))
		}

		stmt := query.Statements[0]
		if !testColumnStatement(t, stmt, tt.expectedIdentifier) {
			return
		}

		val := stmt.(*ast.ColumnStatement).Value
		if !testLiteralExpression(t, val, tt.expectedValue) {
			return
		}
	}

}

func testColumnStatement(t *testing.T, s ast.Statement, name string) bool {
	letStmt, ok := s.(*ast.ColumnStatement)
	if !ok {
		t.Errorf("s not *ast.LetStatement. got=%T", s)
	}
	if letStmt.Name.Value != name {
		t.Errorf("letStmt.Name.Value not '%s'. got=%s", name, letStmt.Name.Value)
		return false
	}

	if letStmt.Name.TokenLiteral() != name {
		t.Errorf("letStmt.Name.TokenLiteral() not '%s'. got=%s",
			name, letStmt.TokenLiteral())
		return false
	}

	return true
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
			"3 + 4, -5 * 5",
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

func TestCallExpressionParsing(t *testing.T) {
	input := "add(1, 2 * 3, 4 + 5)"

	l := lexer.New(input)
	p := New(l)
	query := p.ParseQuery()
	checkParserErrors(t, p)

	if len(query.Statements) != 1 {
		t.Fatalf("query.Statements does not contain %d statements. got=%d\n",
			1, len(query.Statements))
	}

	stmt, ok := query.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("stmt is not ast.ExpressionStatement. got=%T", query.Statements[0])
	}

	exp, ok := stmt.Expression.(*ast.CallExpression)
	if !ok {
		t.Fatalf("stmt.Expression is not ast.CallExpression. got=%T", stmt.Expression)
	}

	if !testIdentifier(t, exp.Function, "add") {
		return
	}

	if len(exp.Arguments) != 3 {
		t.Fatalf("wrong length of arguments. got=%d", len(exp.Arguments))
	}

	testLiteralExpression(t, exp.Arguments[0], 1)
	testInfixExpresssion(t, exp.Arguments[1], 2, "*", 3)
	testInfixExpresssion(t, exp.Arguments[2], 4, "+", 5)
}
