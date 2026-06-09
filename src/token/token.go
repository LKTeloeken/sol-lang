package token

import "fmt"

type Type int

const (
	ILLEGAL Type = iota
	EOF

	// Literals
	INT
	FLOAT
	STRING
	BOOL
	NULL

	// Identifiers
	IDENT

	// Sun-themed keywords
	RISE
	RAY
	GLOW
	RADIATE
	EMIT
	FLARE
	ORBIT
	STAR

	// General keywords
	PUBLIC
	PRIVATE
	VAR
	NEW
	THIS
	IF
	ELSE
	WHILE
	FOR
	EACH
	IN
	BREAK
	CONTINUE
	TRY
	CATCH
	TRUE
	FALSE

	// Type keywords
	INT_TYPE
	FLOAT_TYPE
	BOOL_TYPE
	STRING_TYPE
	VOID_TYPE

	// Operators
	ASSIGN
	PLUS
	MINUS
	ASTERISK
	SLASH
	PERCENT
	BANG
	EQ
	NOT_EQ
	LT
	GT
	LT_EQ
	GT_EQ
	AND
	OR

	// Delimiters
	LPAREN
	RPAREN
	LBRACE
	RBRACE
	LBRACK
	RBRACK
	DOT
	DOTDOT
	COMMA
	SEMICOLON
)

var keywords = map[string]Type{
	"rise":     RISE,
	"ray":      RAY,
	"glow":     GLOW,
	"radiate":  RADIATE,
	"emit":     EMIT,
	"flare":    FLARE,
	"orbit":    ORBIT,
	"star":     STAR,
	"public":   PUBLIC,
	"private":  PRIVATE,
	"var":      VAR,
	"new":      NEW,
	"this":     THIS,
	"if":       IF,
	"else":     ELSE,
	"while":    WHILE,
	"for":      FOR,
	"each":     EACH,
	"in":       IN,
	"break":    BREAK,
	"continue": CONTINUE,
	"try":      TRY,
	"catch":    CATCH,
	"true":     TRUE,
	"false":    FALSE,
	"null":     NULL,
	"int":      INT_TYPE,
	"float":    FLOAT_TYPE,
	"bool":     BOOL_TYPE,
	"string":   STRING_TYPE,
	"void":     VOID_TYPE,
}

func LookupIdent(literal string) Type {
	if tok, ok := keywords[literal]; ok {
		return tok
	}
	return IDENT
}

func (t Type) String() string {
	switch t {
	case ILLEGAL:
		return "ILLEGAL"

	case EOF:
		return "EOF"

	case INT:
		return "INT"

	case FLOAT:
		return "FLOAT"

	case STRING:
		return "STRING"

	case BOOL:
		return "BOOL"

	case NULL:
		return "NULL"

	case IDENT:
		return "IDENT"

	case RISE:
		return "rise"

	case RAY:
		return "ray"

	case GLOW:
		return "glow"

	case RADIATE:
		return "radiate"

	case EMIT:
		return "emit"

	case FLARE:
		return "flare"

	case ORBIT:
		return "orbit"

	case STAR:
		return "star"

	case PUBLIC:
		return "public"

	case PRIVATE:
		return "private"

	case VAR:
		return "var"

	case NEW:
		return "new"

	case THIS:
		return "this"

	case IF:
		return "if"

	case ELSE:
		return "else"

	case WHILE:
		return "while"

	case FOR:
		return "for"

	case EACH:
		return "each"

	case IN:
		return "in"

	case BREAK:
		return "break"

	case CONTINUE:
		return "continue"

	case TRY:
		return "try"

	case CATCH:
		return "catch"

	case TRUE:
		return "true"

	case FALSE:
		return "false"

	case INT_TYPE:
		return "int"

	case FLOAT_TYPE:
		return "float"

	case BOOL_TYPE:
		return "bool"

	case STRING_TYPE:
		return "string"

	case VOID_TYPE:
		return "void"

	case ASSIGN:
		return "="

	case PLUS:
		return "+"

	case MINUS:
		return "-"

	case ASTERISK:
		return "*"

	case SLASH:
		return "/"

	case PERCENT:
		return "%"

	case BANG:
		return "!"

	case EQ:
		return "=="

	case NOT_EQ:
		return "!="

	case LT:
		return "<"

	case GT:
		return ">"

	case LT_EQ:
		return "<="

	case GT_EQ:
		return ">="

	case AND:
		return "&&"

	case OR:
		return "||"

	case LPAREN:
		return "("

	case RPAREN:
		return ")"

	case LBRACE:
		return "{"

	case RBRACE:
		return "}"

	case LBRACK:
		return "["

	case RBRACK:
		return "]"

	case DOT:
		return "."

	case DOTDOT:
		return ".."

	case COMMA:
		return ","

	case SEMICOLON:
		return ";"

	default:
		return fmt.Sprintf("Type(%d)", t)
	}
}

type Token struct {
	Type   Type
	Lexeme string
	Line   int
	Column int
}

func (t Token) String() string {
	if t.Type == EOF {
		return "EOF"
	}

	return fmt.Sprintf("%s(%q)", t.Type, t.Lexeme)
}
