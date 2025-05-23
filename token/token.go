package token

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
	EQ       = "=="
	NOT_EQ   = "!="
	AND      = "&&"
	OR       = "||"

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

	COLUMNCALL = "@"

	FUNCTION = "FUNCTION"
	STRING   = "STRING"

	TRUE  = "TRUE"
	FALSE = "FALSE"
	IF    = "IF"
	ELSE  = "ELSE"
	NULL  = "NULL"
)

var keywords = map[string]TokenType{
	"true":  TRUE,
	"True":  TRUE,
	"TRUE":  TRUE,
	"false": FALSE,
	"False": FALSE,
	"FALSE": FALSE,
	"if":    IF,
	"else":  ELSE,
	"null":  NULL,
	"Null":  NULL,
	"NULL":  NULL,
}

func LookupIdent(ident string) TokenType {
	if tok, ok := keywords[ident]; ok {
		return tok
	}
	return IDENT
}
