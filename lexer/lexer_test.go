package lexer

import (
	"github.com/vamuscari/dyre/token"
	"testing"
)

// example.com/Customers/*/Purchases?Customers.query=Active,Name,Zip:contains(55555)

func TestNextToken(t *testing.T) {

	input := `Active, Name, Zip: contains(55555)`

	tests := []struct {
		expectedType    token.TokenType
		expectedLiteral string
	}{
		{token.IDENT, "Active"},
		{token.COMMA, ","},
		{token.IDENT, "Name"},
		{token.COMMA, ","},
		{token.IDENT, "Zip"},
		{token.COLON, ":"},
		{token.IDENT, "contains"},
		{token.LPAREN, "("},
		{token.INT, "55555"},
		{token.RPAREN, ")"},
		{token.EOF, ""},
	}

	l := New(input)

	for i, tt := range tests {
		tok := l.NextToken()
		if tok.Type != tt.expectedType {
			t.Fatalf("test[%d] - tokentype wrong. expected=%q, got=%q",
				i, tt.expectedType, tok.Type)
		}
		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("test[%d] - literal wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}

// func TestNextToken(t *testing.T) {
//
// 	input := `=+(){},;`
//
// 	tests := []struct {
// 		expectedType    token.TokenType
// 		expectedLiteral string
// 	}{
// 		{token.ASSIGN, "="},
// 		{token.PLUS, "+"},
// 		{token.LPAREN, "("},
// 		{token.RPAREN, ")"},
// 		{token.LBRACE, "{"},
// 		{token.RBRACE, "}"},
// 		{token.COMMA, ","},
// 		{token.SEMICOLON, ";"},
// 		{token.EOF, ""},
// 	}
//
// 	l := New(input)
//
// 	for i, tt := range tests {
// 		tok := l.NextToken()
// 		if tok.Type != tt.expectedType {
// 			t.Fatalf("test[%d] - tokentype wrong. expected=%q, got=%q",
// 				i, tt.expectedType, tok.Type)
// 		}
// 		if tok.Literal != tt.expectedLiteral {
// 			t.Fatalf("test[%d] - literal wrong. expected=%q, got=%q",
// 				i, tt.expectedLiteral, tok.Literal)
// 		}
// 	}
// }
