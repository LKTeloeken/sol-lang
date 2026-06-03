package lexer

import (
	"github.com/unisc/compiladores/sol/internal/token"
)

type Lexer struct {
	input        string
	position     int
	readPosition int
	ch           byte
	line         int
	column       int
}

func New(input string) *Lexer {
	l := &Lexer{input: input, line: 1, column: 1}
	l.readChar()
	return l
}

func (l *Lexer) readChar() {
	if l.readPosition >= len(l.input) {
		l.ch = 0
	} else {
		l.ch = l.input[l.readPosition]
	}
	l.position = l.readPosition
	l.readPosition++
	if l.ch == '\n' {
		l.line++
		l.column = 1
	} else {
		l.column++
	}
}

func (l *Lexer) peekChar() byte {
	if l.readPosition >= len(l.input) {
		return 0
	}
	return l.input[l.readPosition]
}

func (l *Lexer) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\r' || l.ch == '\n' {
		l.readChar()
	}
}

func (l *Lexer) skipComment() bool {
	if l.ch == '/' && l.peekChar() == '/' {
		for l.ch != 0 && l.ch != '\n' {
			l.readChar()
		}
		return true
	}
	if l.ch == '/' && l.peekChar() == '*' {
		l.readChar()
		l.readChar()
		for l.ch != 0 {
			if l.ch == '*' && l.peekChar() == '/' {
				l.readChar()
				l.readChar()
				return true
			}
			l.readChar()
		}
		return true
	}
	return false
}

func (l *Lexer) NextToken() token.Token {
	for {
		l.skipWhitespace()
		for l.skipComment() {
			l.skipWhitespace()
		}

		line := l.line
		col := l.column

		switch l.ch {
		case 0:
			return token.Token{Type: token.EOF, Line: line, Column: col}
		case ';':
			l.readChar()
			return token.Token{Type: token.SEMICOLON, Lexeme: ";", Line: line, Column: col}
		case ',':
			l.readChar()
			return token.Token{Type: token.COMMA, Lexeme: ",", Line: line, Column: col}
		case '.':
			if l.peekChar() == '.' {
				l.readChar()
				l.readChar()
				return token.Token{Type: token.DOTDOT, Lexeme: "..", Line: line, Column: col}
			}
			l.readChar()
			return token.Token{Type: token.DOT, Lexeme: ".", Line: line, Column: col}
		case '(':
			l.readChar()
			return token.Token{Type: token.LPAREN, Lexeme: "(", Line: line, Column: col}
		case ')':
			l.readChar()
			return token.Token{Type: token.RPAREN, Lexeme: ")", Line: line, Column: col}
		case '{':
			l.readChar()
			return token.Token{Type: token.LBRACE, Lexeme: "{", Line: line, Column: col}
		case '}':
			l.readChar()
			return token.Token{Type: token.RBRACE, Lexeme: "}", Line: line, Column: col}
		case '[':
			l.readChar()
			return token.Token{Type: token.LBRACK, Lexeme: "[", Line: line, Column: col}
		case ']':
			l.readChar()
			return token.Token{Type: token.RBRACK, Lexeme: "]", Line: line, Column: col}
		case '+':
			l.readChar()
			return token.Token{Type: token.PLUS, Lexeme: "+", Line: line, Column: col}
		case '-':
			l.readChar()
			return token.Token{Type: token.MINUS, Lexeme: "-", Line: line, Column: col}
		case '*':
			l.readChar()
			return token.Token{Type: token.ASTERISK, Lexeme: "*", Line: line, Column: col}
		case '/':
			l.readChar()
			return token.Token{Type: token.SLASH, Lexeme: "/", Line: line, Column: col}
		case '%':
			l.readChar()
			return token.Token{Type: token.PERCENT, Lexeme: "%", Line: line, Column: col}
		case '!':
			if l.peekChar() == '=' {
				l.readChar()
				l.readChar()
				return token.Token{Type: token.NOT_EQ, Lexeme: "!=", Line: line, Column: col}
			}
			l.readChar()
			return token.Token{Type: token.BANG, Lexeme: "!", Line: line, Column: col}
		case '=':
			if l.peekChar() == '=' {
				l.readChar()
				l.readChar()
				return token.Token{Type: token.EQ, Lexeme: "==", Line: line, Column: col}
			}
			l.readChar()
			return token.Token{Type: token.ASSIGN, Lexeme: "=", Line: line, Column: col}
		case '<':
			if l.peekChar() == '=' {
				l.readChar()
				l.readChar()
				return token.Token{Type: token.LT_EQ, Lexeme: "<=", Line: line, Column: col}
			}
			l.readChar()
			return token.Token{Type: token.LT, Lexeme: "<", Line: line, Column: col}
		case '>':
			if l.peekChar() == '=' {
				l.readChar()
				l.readChar()
				return token.Token{Type: token.GT_EQ, Lexeme: ">=", Line: line, Column: col}
			}
			l.readChar()
			return token.Token{Type: token.GT, Lexeme: ">", Line: line, Column: col}
		case '&':
			if l.peekChar() == '&' {
				l.readChar()
				l.readChar()
				return token.Token{Type: token.AND, Lexeme: "&&", Line: line, Column: col}
			}
			return token.Token{Type: token.ILLEGAL, Lexeme: string(l.ch), Line: line, Column: col}
		case '|':
			if l.peekChar() == '|' {
				l.readChar()
				l.readChar()
				return token.Token{Type: token.OR, Lexeme: "||", Line: line, Column: col}
			}
			return token.Token{Type: token.ILLEGAL, Lexeme: string(l.ch), Line: line, Column: col}
		case '"':
			return l.readString(line, col)
		default:
			if isLetter(l.ch) {
				return l.readIdentifier(line, col)
			}
			if isDigit(l.ch) {
				return l.readNumber(line, col)
			}
			tok := token.Token{Type: token.ILLEGAL, Lexeme: string(l.ch), Line: line, Column: col}
			l.readChar()
			return tok
		}
	}
}

func (l *Lexer) readIdentifier(line, col int) token.Token {
	start := l.position
	for isLetter(l.ch) || isDigit(l.ch) {
		l.readChar()
	}
	lit := l.input[start:l.position]
	typ := token.LookupIdent(lit)
	if typ == token.TRUE || typ == token.FALSE {
		return token.Token{Type: token.BOOL, Lexeme: lit, Line: line, Column: col}
	}
	if typ == token.NULL {
		return token.Token{Type: token.NULL, Lexeme: lit, Line: line, Column: col}
	}
	return token.Token{Type: typ, Lexeme: lit, Line: line, Column: col}
}

func (l *Lexer) readNumber(line, col int) token.Token {
	start := l.position
	isFloat := false
	for isDigit(l.ch) {
		l.readChar()
	}
	if l.ch == '.' && isDigit(l.peekChar()) {
		isFloat = true
		l.readChar()
		for isDigit(l.ch) {
			l.readChar()
		}
	}
	lit := l.input[start:l.position]
	typ := token.INT
	if isFloat {
		typ = token.FLOAT
	}
	return token.Token{Type: typ, Lexeme: lit, Line: line, Column: col}
}

func (l *Lexer) readString(line, col int) token.Token {
	l.readChar()
	start := l.position
	for l.ch != 0 && l.ch != '"' {
		if l.ch == '\\' {
			l.readChar()
		}
		l.readChar()
	}
	if l.ch != '"' {
		return token.Token{Type: token.ILLEGAL, Lexeme: "unterminated string", Line: line, Column: col}
	}
	lit := l.input[start:l.position]
	l.readChar()
	return token.Token{Type: token.STRING, Lexeme: lit, Line: line, Column: col}
}

func isLetter(ch byte) bool {
	return ch == '_' || (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z')
}

func isDigit(ch byte) bool {
	return ch >= '0' && ch <= '9'
}

// Tokenize reads all tokens until EOF.
func Tokenize(input string) []token.Token {
	l := New(input)
	var tokens []token.Token
	for {
		tok := l.NextToken()
		tokens = append(tokens, tok)
		if tok.Type == token.EOF {
			break
		}
	}
	return tokens
}
