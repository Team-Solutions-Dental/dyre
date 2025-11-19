package parser

import (
	"fmt"
	"strconv"

	"github.com/Team-Solutions-Dental/dyre/ast"
	"github.com/Team-Solutions-Dental/dyre/lexer"
	"github.com/Team-Solutions-Dental/dyre/token"
)

const (
	_ int = iota
	LOWEST
	CONDITION   // AND OR
	EQUALS      // ==
	LESSGREATER // > or <
	SUM         // +
	PRODUCT     // *
	PREFIX      // -X or !X
	CALL        // myFunction(X)
)

var precedences = map[token.TokenType]int{
	token.AND:      CONDITION,
	token.OR:       CONDITION,
	token.EQ:       EQUALS,
	token.NOT_EQ:   EQUALS,
	token.LT:       LESSGREATER,
	token.LTE:      LESSGREATER,
	token.GT:       LESSGREATER,
	token.GTE:      LESSGREATER,
	token.PLUS:     SUM,
	token.MINUS:    SUM,
	token.SLASH:    PRODUCT,
	token.ASTERISK: PRODUCT,
	token.LPAREN:   CALL,
}

type Parser struct {
	l *lexer.Lexer

	curToken  token.Token
	peekToken token.Token

	parserErrors []string

	prefixParseFns map[token.TokenType]prefixParseFn
	infixParseFns  map[token.TokenType]infixParseFn
}

func New(l *lexer.Lexer) *Parser {
	p := &Parser{
		l:            l,
		parserErrors: []string{},
	}

	//Read two tokens for init cur and peek
	p.nextToken()
	p.nextToken()

	p.prefixParseFns = make(map[token.TokenType]prefixParseFn)
	p.registerPrefix(token.IDENT, p.parseIdentifier)
	p.registerPrefix(token.NULL, p.parseNullLiteral)
	p.registerPrefix(token.INT, p.parseIntegerLiteral)
	p.registerPrefix(token.BANG, p.parsePrefixExpression)
	p.registerPrefix(token.MINUS, p.parsePrefixExpression)
	p.registerPrefix(token.TRUE, p.parseBoolean)
	p.registerPrefix(token.FALSE, p.parseBoolean)
	p.registerPrefix(token.LPAREN, p.parseGroupedExpression)
	p.registerPrefix(token.EQ, p.parseColumnPrefixExpression)
	p.registerPrefix(token.NOT_EQ, p.parseColumnPrefixExpression)
	p.registerPrefix(token.LT, p.parseColumnPrefixExpression)
	p.registerPrefix(token.LTE, p.parseColumnPrefixExpression)
	p.registerPrefix(token.GT, p.parseColumnPrefixExpression)
	p.registerPrefix(token.GTE, p.parseColumnPrefixExpression)
	p.registerPrefix(token.REFERENCE, p.parseReference)
	p.registerPrefix(token.STRING, p.parseStringLiteral)
	p.registerPrefix(token.ASC, p.parseOrderExpression)
	p.registerPrefix(token.DESC, p.parseOrderExpression)

	p.infixParseFns = make(map[token.TokenType]infixParseFn)
	p.registerInfix(token.PLUS, p.parseInfixExpression)
	p.registerInfix(token.MINUS, p.parseInfixExpression)
	p.registerInfix(token.SLASH, p.parseInfixExpression)
	p.registerInfix(token.ASTERISK, p.parseInfixExpression)
	p.registerInfix(token.EQ, p.parseInfixExpression)
	p.registerInfix(token.NOT_EQ, p.parseInfixExpression)
	p.registerInfix(token.LT, p.parseInfixExpression)
	p.registerInfix(token.LTE, p.parseInfixExpression)
	p.registerInfix(token.GT, p.parseInfixExpression)
	p.registerInfix(token.GTE, p.parseInfixExpression)
	p.registerInfix(token.AND, p.parseInfixExpression)
	p.registerInfix(token.OR, p.parseInfixExpression)
	p.registerInfix(token.LPAREN, p.parseCallExpression)

	return p
}

func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.l.NextToken()
}

func (p *Parser) ParseQuery() *ast.RequestStatements {
	query := &ast.RequestStatements{}
	query.Statements = []ast.Statement{}

	for p.curToken.Type != token.EOF {
		col := p.parseStatement()
		if col != nil {
			query.Statements = append(query.Statements, col)
		}
		p.nextToken()
	}
	return query
}

func (p *Parser) parseStatement() ast.Statement {
	switch {
	case p.curTokenIs(token.IDENT) && p.peekTokenIs(token.COLON):
		return p.parseColumnLiteral()
	case p.curTokenIs(token.COLUMN):
		return p.parseColumnFunction()
	case p.curTokenIs(token.GROUP):
		return p.parseGroupFunction()
	default:
		return p.parseExpressionStatement()
	}
}

// Expects ident, colon
// Since the transpiler maintains a ref to the current column
// they will no longer hold reference following expressions
func (p *Parser) parseColumnLiteral() *ast.ColumnLiteral {
	lit := &ast.ColumnLiteral{Token: p.curToken}

	p.nextToken()

	for p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return lit
}

func (p *Parser) parseColumnFunction() *ast.ColumnFunction {
	funct := &ast.ColumnFunction{Token: p.curToken, Fn: p.curToken.Literal}

	if !p.peekTokenIs(token.LPAREN) {
		p.peekError(token.LPAREN)
		return nil
	}

	p.nextToken()

	funct.Arguments = p.parseCallArguments()

	if !p.peekTokenIs(token.COLON) {
		p.peekError(token.COLON)
		return nil
	}

	p.nextToken()

	return funct
}

func (p *Parser) parseGroupFunction() *ast.GroupFunction {
	funct := &ast.GroupFunction{Token: p.curToken, Fn: p.curToken.Literal}

	p.nextToken()

	funct.Arguments = p.parseCallArguments()

	if !p.peekTokenIs(token.COLON) {
		p.peekError(token.COLON)
		return nil
	}

	p.nextToken()

	return funct
}

func (p *Parser) parseIdentifier() ast.Expression {
	return &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
}

func (p *Parser) parseExpressionStatement() ast.Statement {
	stmt := &ast.ExpressionStatement{Token: p.curToken}

	if !p.curTokenIs(token.SEMICOLON) {
		stmt.Expression = p.parseExpression(LOWEST)
	}

	for p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseGroupedExpression() ast.Expression {
	p.nextToken()

	exp := p.parseExpression(LOWEST)

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	return exp
}

func (p *Parser) noPrefixParseFnError(t token.TokenType) {
	msg := fmt.Sprintf("no prefix parse function for %s found", t)
	p.parserErrors = append(p.parserErrors, msg)
}

func (p *Parser) parseExpression(precedence int) ast.Expression {
	prefix := p.prefixParseFns[p.curToken.Type]
	if prefix == nil {
		p.noPrefixParseFnError(p.curToken.Type)
		return nil
	}

	leftExp := prefix()

	for !p.peekTokenIs(token.SEMICOLON) && !p.peekTokenIs(token.EOF) && precedence < p.peekPrecedence() {
		infix := p.infixParseFns[p.peekToken.Type]
		if infix == nil {
			return leftExp
		}

		p.nextToken()

		leftExp = infix(leftExp)
	}

	return leftExp
}

func (p *Parser) parsePrefixExpression() ast.Expression {
	expression := &ast.PrefixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
	}

	p.nextToken()

	expression.Right = p.parseExpression(PREFIX)

	return expression
}

func (p *Parser) parseInfixExpression(left ast.Expression) ast.Expression {
	expression := &ast.InfixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
		Left:     left,
	}

	precedence := p.curPrecedence()
	p.nextToken()
	expression.Right = p.parseExpression(precedence)

	return expression
}

func (p *Parser) peekPrecedence() int {
	if p, ok := precedences[p.peekToken.Type]; ok {
		return p
	}

	return LOWEST
}

func (p *Parser) curPrecedence() int {
	if p, ok := precedences[p.curToken.Type]; ok {
		return p
	}

	return LOWEST
}

func (p *Parser) curTokenIs(t token.TokenType) bool {
	return p.curToken.Type == t
}

func (p *Parser) peekTokenIs(t token.TokenType) bool {
	return p.peekToken.Type == t
}

func (p *Parser) expectPeek(t token.TokenType) bool {
	if p.peekTokenIs(t) {
		p.nextToken()
		return true
	} else {
		p.peekError(t)
		return false
	}
}

func (p *Parser) Errors() []string {
	return p.parserErrors
}

func (p *Parser) peekError(t token.TokenType) {
	msg := fmt.Sprintf("expected next token to be %s, got %s instead", t, p.peekToken.Type)
	p.parserErrors = append(p.parserErrors, msg)
}

type (
	prefixParseFn func() ast.Expression
	infixParseFn  func(ast.Expression) ast.Expression
)

func (p *Parser) registerPrefix(tokenType token.TokenType, fn prefixParseFn) {
	p.prefixParseFns[tokenType] = fn
}

func (p *Parser) registerInfix(tokenType token.TokenType, fn infixParseFn) {
	p.infixParseFns[tokenType] = fn
}

func (p *Parser) parseIntegerLiteral() ast.Expression {
	lit := &ast.IntegerLiteral{Token: p.curToken}

	value, err := strconv.ParseInt(p.curToken.Literal, 0, 64)
	if err != nil {
		msg := fmt.Sprintf("could not parse %q as integer", p.curToken.Literal)
		p.parserErrors = append(p.parserErrors, msg)
		return nil
	}

	lit.Value = value

	return lit
}

func (p *Parser) parseBoolean() ast.Expression {
	return &ast.Boolean{Token: p.curToken, Value: p.curTokenIs(token.TRUE)}
}

func (p *Parser) parseCallExpression(function ast.Expression) ast.Expression {
	exp := &ast.CallExpression{Token: p.curToken, Function: function}
	exp.Arguments = p.parseCallArguments()
	return exp
}

func (p *Parser) parseCallArguments() []ast.Expression {
	args := []ast.Expression{}

	if p.peekTokenIs(token.RPAREN) {
		p.nextToken()
		return args
	}

	p.nextToken()
	args = append(args, p.parseExpression(LOWEST))

	for p.peekTokenIs(token.COMMA) {
		p.nextToken()
		p.nextToken()
		args = append(args, p.parseExpression(LOWEST))
	}

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	return args
}

// parse @ or @(arg) where arg is string
func (p *Parser) parseReference() ast.Expression {
	ref := &ast.Reference{Token: p.curToken}

	if !p.peekTokenIs(token.LPAREN) {
		return ref
	}

	p.nextToken()
	p.nextToken()

	ref.Argument = p.parseStringLiteral()

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	return ref

}

// Casts a prefix expression as an infix expression assuming a column is being referenced
// Col: != NULL -> Col: @ != NULL
func (p *Parser) parseColumnPrefixExpression() ast.Expression {
	expression := &ast.InfixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
		Left:     &ast.Reference{Token: token.Token{Type: token.REFERENCE, Literal: "@"}},
	}

	precedence := p.curPrecedence()
	p.nextToken()
	expression.Right = p.parseExpression(precedence)

	return expression
}

func (p *Parser) parseNullLiteral() ast.Expression {
	return &ast.NullLiteral{Token: p.curToken}
}

func (p *Parser) parseStringLiteral() ast.Expression {
	return &ast.StringLiteral{Token: p.curToken, Value: p.curToken.Literal}
}

func (p *Parser) parseOrderExpression() ast.Expression {
	asc := true
	if p.curTokenIs(token.DESC) {
		asc = false
	}
	return &ast.OrderExpression{Token: p.curToken, Ascending: asc}
}
