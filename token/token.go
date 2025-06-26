package token

import (
	"strings"
)

type TokenType string
type Token struct {
	Type    TokenType
	Literal string
}

const (
	ILLEGAL = "ILLEGAL"
	EOF     = "EOF"

	IDENT  = "IDENT"
	INT    = "INT"
	COLUMN = "COLUMN"

	// Operations
	ASSIGN   = "="
	PLUS     = "+"
	MINUS    = "-"
	BANG     = "!"
	ASTERISK = "*"
	SLASH    = "/"
	BKSLASH  = "\\"
	EQ       = "=="
	NOT_EQ   = "!="

	LT  = "<"
	GT  = ">"
	LTE = "<="
	GTE = ">="

	COMMA     = ","
	SEMICOLON = ";"
	COLON     = ":"

	LPAREN = "("
	RPAREN = ")"
	LBRACE = "}"
	RBRACE = "{"

	FUNCTION = "FUNCTION"
	STRING   = "STRING"

	TRUE  = "TRUE"
	FALSE = "FALSE"
	AND   = "AND"
	OR    = "OR"
	NULL  = "NULL"

	REFERENCE = "@"

	GROUP = "GROUP"

	ASC  = "ASC"
	DESC = "DESC"
)

var keywords = map[string]TokenType{
	"TRUE":  TRUE,
	"FALSE": FALSE,
	"NULL":  NULL,
	"AND":   AND,
	"OR":    OR,
	"ASC":   ASC,
	"DESC":  DESC,
	// Column Functions
	"AS":      COLUMN,
	"ALIAS":   COLUMN,
	"EXCLUDE": COLUMN,
	// Group Functions
	"GROUP": GROUP,
	"COUNT": GROUP,
	"AVG":   GROUP,
	"SUM":   GROUP,
	"MIN":   GROUP,
	"MAX":   GROUP,
}

func LookupIdent(ident string) TokenType {
	if tok, ok := keywords[strings.ToUpper(ident)]; ok {
		return tok
	}
	return IDENT
}
