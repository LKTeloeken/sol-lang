package parser

import (
	"github.com/alecthomas/participle/v2"

	"github.com/unisc/compiladores/sol/src/ast"
	"github.com/unisc/compiladores/sol/src/diag"
	"github.com/unisc/compiladores/sol/src/lexer"
)

var solParser = participle.MustBuild[PProgram](
	participle.Lexer(lexer.Def),
	participle.Elide("Whitespace", "Comment"),
	participle.UseLookahead(participle.MaxLookahead),
	participle.Unquote("String"),
)

type Parser struct {
	src    string
	file   string
	errors []diag.Error
}

// New creates a parser for src. The file parameter is used in error messages.
func New(src, file string) *Parser {
	return &Parser{src: src, file: file}
}

func (p *Parser) Errors() []diag.Error { return p.errors }

func (p *Parser) Parse() *ast.Program {
	prog, err := solParser.ParseString(p.file, p.src)
	if err != nil {
		p.errors = append(p.errors, diagFromError(p.file, err))
		return &ast.Program{}
	}
	return convertProgram(prog)
}

func diagFromError(file string, err error) diag.Error {
	if pe, ok := err.(participle.Error); ok {
		return diag.Error{
			File:    file,
			Line:    pe.Position().Line,
			Column:  pe.Position().Column,
			Message: pe.Message(),
		}
	}
	return diag.Error{File: file, Message: err.Error()}
}
