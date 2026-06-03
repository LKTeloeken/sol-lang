package parser

import (
	"fmt"
	"strconv"

	"github.com/unisc/compiladores/sol/internal/ast"
	"github.com/unisc/compiladores/sol/internal/diag"
	"github.com/unisc/compiladores/sol/internal/lexer"
	"github.com/unisc/compiladores/sol/internal/token"
)

type Parser struct {
	l      *lexer.Lexer
	file   string
	cur    token.Token
	peek   token.Token
	errors []diag.Error
}

func New(l *lexer.Lexer, file string) *Parser {
	p := &Parser{l: l, file: file}
	p.cur, p.peek = p.l.NextToken(), p.l.NextToken()
	return p
}

func (p *Parser) Errors() []diag.Error { return p.errors }

func (p *Parser) Parse() *ast.Program {
	prog := &ast.Program{}
	for p.cur.Type != token.EOF {
		if p.cur.Type == token.RISE {
			prog.Decls = append(prog.Decls, p.parseClassDecl())
		} else {
			prog.Decls = append(prog.Decls, p.parseTopLevelStmt())
		}
	}
	return prog
}

func (p *Parser) advance() {
	p.cur, p.peek = p.peek, p.l.NextToken()
}

func (p *Parser) pos() ast.Pos {
	return ast.Pos{Line: p.cur.Line, Column: p.cur.Column}
}

func (p *Parser) errorf(format string, args ...any) {
	p.errors = append(p.errors, diag.Error{
		File: p.file, Line: p.cur.Line, Column: p.cur.Column,
		Message: fmt.Sprintf(format, args...),
	})
}

func (p *Parser) expect(typ token.Type) bool {
	if p.cur.Type == typ {
		p.advance()
		return true
	}
	p.errorf("expected %s, got %s", typ, p.cur.Type)
	return false
}

func (p *Parser) parseClassDecl() *ast.ClassDecl {
	pos := p.pos()
	p.expect(token.RISE)
	name := p.cur.Lexeme
	if !p.expect(token.IDENT) {
		return &ast.ClassDecl{PosInfo: pos, Name: name}
	}
	super := ""
	if p.cur.Type == token.ENLIGHTS {
		p.advance()
		super = p.cur.Lexeme
		p.expect(token.IDENT)
	}
	p.expect(token.LBRACE)
	var members []ast.MemberDecl
	for p.cur.Type != token.RBRACE && p.cur.Type != token.EOF {
		members = append(members, p.parseMemberDecl())
	}
	p.expect(token.RBRACE)
	return &ast.ClassDecl{PosInfo: pos, Name: name, SuperName: super, Members: members}
}

func (p *Parser) parseMemberDecl() ast.MemberDecl {
	switch p.cur.Type {
	case token.PUBLIC, token.PRIVATE:
		vis := p.cur.Lexeme
		p.advance()
		if p.cur.Type == token.RAY {
			return p.parseRayDecl(vis)
		}
		return p.parseFieldDecl(vis)
	case token.GLOW:
		return p.parseGlowDecl()
	default:
		p.errorf("unexpected token in class body: %s", p.cur.Type)
		p.advance()
		return nil
	}
}

func (p *Parser) parseFieldDecl(vis string) *ast.FieldDecl {
	pos := p.pos()
	typ := p.parseType()
	name := p.cur.Lexeme
	p.expect(token.IDENT)
	p.expect(token.SEMICOLON)
	return &ast.FieldDecl{PosInfo: pos, Visibility: vis, Type: typ, Name: name}
}

func (p *Parser) parseGlowDecl() *ast.GlowDecl {
	pos := p.pos()
	p.expect(token.GLOW)
	p.expect(token.LPAREN)
	params := p.parseParamList()
	p.expect(token.RPAREN)
	body := p.parseBlock()
	return &ast.GlowDecl{PosInfo: pos, Params: params, Body: body}
}

func (p *Parser) parseRayDecl(vis string) *ast.RayDecl {
	pos := p.pos()
	p.expect(token.RAY)
	name := p.cur.Lexeme
	p.expect(token.IDENT)
	p.expect(token.LPAREN)
	params := p.parseParamList()
	p.expect(token.RPAREN)
	var retType *ast.TypeDesc
	if p.isTypeStart() {
		retType = p.parseType()
	}
	body := p.parseBlock()
	return &ast.RayDecl{PosInfo: pos, Visibility: vis, Name: name, Params: params, ReturnType: retType, Body: body}
}

func (p *Parser) parseParamList() []ast.Param {
	var params []ast.Param
	if p.cur.Type == token.RPAREN {
		return params
	}
	for {
		pos := p.pos()
		typ := p.parseType()
		name := p.cur.Lexeme
		p.expect(token.IDENT)
		params = append(params, ast.Param{Type: typ, Name: name, Pos: pos})
		if p.cur.Type != token.COMMA {
			break
		}
		p.advance()
	}
	return params
}

func (p *Parser) isTypeStart() bool {
	switch p.cur.Type {
	case token.INT_TYPE, token.FLOAT_TYPE, token.BOOL_TYPE, token.STRING_TYPE, token.VOID_TYPE, token.LBRACK, token.IDENT:
		return true
	default:
		return false
	}
}

func (p *Parser) parseType() *ast.TypeDesc {
	pos := p.pos()
	if p.cur.Type == token.LBRACK {
		p.advance()
		elem := p.parseType()
		p.expect(token.RBRACK)
		return &ast.TypeDesc{Pos: pos, IsArray: true, ElemType: elem}
	}
	base := p.cur.Lexeme
	switch p.cur.Type {
	case token.INT_TYPE, token.FLOAT_TYPE, token.BOOL_TYPE, token.STRING_TYPE, token.VOID_TYPE:
		p.advance()
	case token.IDENT:
		p.advance()
	default:
		p.errorf("expected type, got %s", p.cur.Type)
		return &ast.TypeDesc{Pos: pos, Base: "void"}
	}
	return &ast.TypeDesc{Pos: pos, Base: base}
}

func (p *Parser) parseBlock() *ast.BlockStmt {
	pos := p.pos()
	p.expect(token.LBRACE)
	var stmts []ast.Stmt
	for p.cur.Type != token.RBRACE && p.cur.Type != token.EOF {
		stmts = append(stmts, p.parseStatement())
	}
	p.expect(token.RBRACE)
	return &ast.BlockStmt{PosInfo: pos, Stmts: stmts}
}

func (p *Parser) parseTopLevelStmt() ast.TopLevelDecl {
	return p.parseStatement().(ast.TopLevelDecl)
}

func (p *Parser) parseStatement() ast.Stmt {
	switch p.cur.Type {
	case token.VAR:
		return p.parseVarDecl()
	case token.IF:
		return p.parseIfStmt()
	case token.WHILE:
		return p.parseWhileStmt()
	case token.FOR:
		p.advance()
		if p.cur.Type == token.EACH {
			return p.parseForEachStmt()
		}
		return p.parseForRangeStmt()
	case token.BREAK:
		return p.parseBreakStmt()
	case token.CONTINUE:
		return p.parseContinueStmt()
	case token.EMIT:
		return p.parseEmitStmt()
	case token.FLARE:
		return p.parseFlareStmt()
	case token.TRY:
		return p.parseTryCatchStmt()
	case token.LBRACE:
		return p.parseBlock()
	default:
		return p.parseAssignOrExprStmt()
	}
}

func (p *Parser) parseVarDecl() *ast.VarDeclStmt {
	pos := p.pos()
	p.expect(token.VAR)
	name := p.cur.Lexeme
	p.expect(token.IDENT)
	typ := p.parseType()
	var val ast.Expr
	if p.cur.Type == token.ASSIGN {
		p.advance()
		val = p.parseExpression()
	}
	p.expect(token.SEMICOLON)
	return &ast.VarDeclStmt{PosInfo: pos, Name: name, Type: typ, Value: val}
}

func (p *Parser) parseAssignOrExprStmt() ast.Stmt {
	pos := p.pos()
	expr := p.parseExpression()
	if p.cur.Type == token.ASSIGN {
		p.advance()
		val := p.parseExpression()
		p.expect(token.SEMICOLON)
		return &ast.AssignStmt{PosInfo: pos, Target: expr, Value: val}
	}
	p.expect(token.SEMICOLON)
	return &ast.ExprStmt{PosInfo: pos, Expr: expr}
}

func (p *Parser) parseIfStmt() *ast.IfStmt {
	pos := p.pos()
	p.expect(token.IF)
	p.expect(token.LPAREN)
	cond := p.parseExpression()
	p.expect(token.RPAREN)
	then := p.parseBlock()
	var els *ast.BlockStmt
	if p.cur.Type == token.ELSE {
		p.advance()
		els = p.parseBlock()
	}
	return &ast.IfStmt{PosInfo: pos, Cond: cond, Then: then, Else: els}
}

func (p *Parser) parseWhileStmt() *ast.WhileStmt {
	pos := p.pos()
	p.expect(token.WHILE)
	p.expect(token.LPAREN)
	cond := p.parseExpression()
	p.expect(token.RPAREN)
	body := p.parseBlock()
	return &ast.WhileStmt{PosInfo: pos, Cond: cond, Body: body}
}

func (p *Parser) parseForEachStmt() *ast.ForEachStmt {
	pos := p.pos()
	p.expect(token.EACH)
	name := p.cur.Lexeme
	p.expect(token.IDENT)
	p.expect(token.IN)
	iter := p.parseExpression()
	body := p.parseBlock()
	return &ast.ForEachStmt{PosInfo: pos, VarName: name, Iter: iter, Body: body}
}

func (p *Parser) parseForRangeStmt() *ast.ForRangeStmt {
	pos := p.pos()
	name := p.cur.Lexeme
	p.expect(token.IDENT)
	p.expect(token.IN)
	start := p.parseExpression()
	p.expect(token.DOTDOT)
	end := p.parseExpression()
	body := p.parseBlock()
	return &ast.ForRangeStmt{PosInfo: pos, VarName: name, Start: start, End: end, Body: body}
}

func (p *Parser) parseBreakStmt() *ast.BreakStmt {
	pos := p.pos()
	p.expect(token.BREAK)
	p.expect(token.SEMICOLON)
	return &ast.BreakStmt{PosInfo: pos}
}

func (p *Parser) parseContinueStmt() *ast.ContinueStmt {
	pos := p.pos()
	p.expect(token.CONTINUE)
	p.expect(token.SEMICOLON)
	return &ast.ContinueStmt{PosInfo: pos}
}

func (p *Parser) parseEmitStmt() *ast.EmitStmt {
	pos := p.pos()
	p.expect(token.EMIT)
	var val ast.Expr
	if p.cur.Type != token.SEMICOLON {
		val = p.parseExpression()
	}
	p.expect(token.SEMICOLON)
	return &ast.EmitStmt{PosInfo: pos, Value: val}
}

func (p *Parser) parseFlareStmt() *ast.FlareStmt {
	pos := p.pos()
	p.expect(token.FLARE)
	val := p.parseExpression()
	p.expect(token.SEMICOLON)
	return &ast.FlareStmt{PosInfo: pos, Value: val}
}

func (p *Parser) parseTryCatchStmt() *ast.TryCatchStmt {
	pos := p.pos()
	p.expect(token.TRY)
	tryBlock := p.parseBlock()
	p.expect(token.CATCH)
	p.expect(token.LPAREN)
	catchVar := p.cur.Lexeme
	p.expect(token.IDENT)
	p.expect(token.RPAREN)
	catchBlock := p.parseBlock()
	return &ast.TryCatchStmt{PosInfo: pos, Try: tryBlock, CatchVar: catchVar, Catch: catchBlock}
}

// Expression parsing (Pratt / precedence climbing)

func (p *Parser) parseExpression() ast.Expr {
	return p.parseLogicalOr()
}

func (p *Parser) parseLogicalOr() ast.Expr {
	left := p.parseLogicalAnd()
	for p.cur.Type == token.OR {
		op := p.cur.Type
		pos := p.pos()
		p.advance()
		right := p.parseLogicalAnd()
		left = &ast.BinaryExpr{BaseExpr: ast.NewBase(pos), Left: left, Operator: op, Right: right}
	}
	return left
}

func (p *Parser) parseLogicalAnd() ast.Expr {
	left := p.parseEquality()
	for p.cur.Type == token.AND {
		op := p.cur.Type
		pos := p.pos()
		p.advance()
		right := p.parseEquality()
		left = &ast.BinaryExpr{BaseExpr: ast.NewBase(pos), Left: left, Operator: op, Right: right}
	}
	return left
}

func (p *Parser) parseEquality() ast.Expr {
	left := p.parseRelational()
	for p.cur.Type == token.EQ || p.cur.Type == token.NOT_EQ {
		op := p.cur.Type
		pos := p.pos()
		p.advance()
		right := p.parseRelational()
		left = &ast.BinaryExpr{BaseExpr: ast.NewBase(pos), Left: left, Operator: op, Right: right}
	}
	return left
}

func (p *Parser) parseRelational() ast.Expr {
	left := p.parseAdditive()
	for p.cur.Type == token.LT || p.cur.Type == token.GT || p.cur.Type == token.LT_EQ || p.cur.Type == token.GT_EQ {
		op := p.cur.Type
		pos := p.pos()
		p.advance()
		right := p.parseAdditive()
		left = &ast.BinaryExpr{BaseExpr: ast.NewBase(pos), Left: left, Operator: op, Right: right}
	}
	return left
}

func (p *Parser) parseAdditive() ast.Expr {
	left := p.parseMultiplicative()
	for p.cur.Type == token.PLUS || p.cur.Type == token.MINUS {
		op := p.cur.Type
		pos := p.pos()
		p.advance()
		right := p.parseMultiplicative()
		left = &ast.BinaryExpr{BaseExpr: ast.NewBase(pos), Left: left, Operator: op, Right: right}
	}
	return left
}

func (p *Parser) parseMultiplicative() ast.Expr {
	left := p.parseUnary()
	for p.cur.Type == token.ASTERISK || p.cur.Type == token.SLASH || p.cur.Type == token.PERCENT {
		op := p.cur.Type
		pos := p.pos()
		p.advance()
		right := p.parseUnary()
		left = &ast.BinaryExpr{BaseExpr: ast.NewBase(pos), Left: left, Operator: op, Right: right}
	}
	return left
}

func (p *Parser) parseUnary() ast.Expr {
	if p.cur.Type == token.BANG || p.cur.Type == token.MINUS {
		op := p.cur.Type
		pos := p.pos()
		p.advance()
		operand := p.parseUnary()
		return &ast.UnaryExpr{BaseExpr: ast.NewBase(pos), Operator: op, Operand: operand}
	}
	return p.parsePostfix()
}

func (p *Parser) parsePostfix() ast.Expr {
	expr := p.parsePrimary()
	for {
		switch p.cur.Type {
		case token.DOT:
			p.advance()
			field := p.cur.Lexeme
			p.expect(token.IDENT)
			if p.cur.Type == token.LPAREN {
				expr = p.finishCall(expr, field)
			} else {
				expr = &ast.GetFieldExpr{BaseExpr: ast.NewBase(expr.Pos()), Object: expr, Field: field}
			}
		case token.LPAREN:
			if call, ok := expr.(*ast.IdentExpr); ok {
				expr = p.finishCallIdent(call)
			} else {
				return expr
			}
		case token.LBRACK:
			p.advance()
			index := p.parseExpression()
			p.expect(token.RBRACK)
			expr = &ast.IndexExpr{BaseExpr: ast.NewBase(expr.Pos()), Object: expr, Index: index}
		default:
			return expr
		}
	}
}

func (p *Parser) finishCall(callee ast.Expr, method string) ast.Expr {
	pos := callee.Pos()
	p.expect(token.LPAREN)
	args := p.parseArgList()
	p.expect(token.RPAREN)
	return &ast.CallExpr{BaseExpr: ast.NewBase(pos), Callee: &ast.GetFieldExpr{BaseExpr: ast.NewBase(pos), Object: callee, Field: method}, Args: args}
}

func (p *Parser) finishCallIdent(id *ast.IdentExpr) ast.Expr {
	pos := id.Pos()
	p.expect(token.LPAREN)
	args := p.parseArgList()
	p.expect(token.RPAREN)
	return &ast.CallExpr{BaseExpr: ast.NewBase(pos), Callee: id, Args: args}
}

func (p *Parser) parseArgList() []ast.Expr {
	var args []ast.Expr
	if p.cur.Type == token.RPAREN {
		return args
	}
	for {
		args = append(args, p.parseExpression())
		if p.cur.Type != token.COMMA {
			break
		}
		p.advance()
	}
	return args
}

func (p *Parser) parseMemberName() string {
	switch p.cur.Type {
	case token.IDENT, token.GLOW:
		name := p.cur.Lexeme
		p.advance()
		return name
	default:
		p.errorf("expected member name, got %s", p.cur.Type)
		return ""
	}
}

func (p *Parser) parsePrimary() ast.Expr {
	pos := p.pos()
	switch p.cur.Type {
	case token.INT:
		v, _ := strconv.ParseInt(p.cur.Lexeme, 10, 64)
		p.advance()
		return &ast.IntLit{BaseExpr: ast.NewBase(pos), Value: v}
	case token.FLOAT:
		v, _ := strconv.ParseFloat(p.cur.Lexeme, 64)
		p.advance()
		return &ast.FloatLit{BaseExpr: ast.NewBase(pos), Value: v}
	case token.STRING:
		lit := p.cur.Lexeme
		p.advance()
		return &ast.StringLit{BaseExpr: ast.NewBase(pos), Value: lit}
	case token.BOOL:
		v := p.cur.Lexeme == "true"
		p.advance()
		return &ast.BoolLit{BaseExpr: ast.NewBase(pos), Value: v}
	case token.NULL:
		p.advance()
		return &ast.NullLit{BaseExpr: ast.NewBase(pos)}
	case token.THIS:
		p.advance()
		return &ast.ThisExpr{BaseExpr: ast.NewBase(pos)}
	case token.ENLIGHTS:
		p.advance()
		if p.cur.Type == token.DOT {
			p.advance()
			method := p.parseMemberName()
			if p.cur.Type == token.LPAREN {
				p.expect(token.LPAREN)
				args := p.parseArgList()
				p.expect(token.RPAREN)
				return &ast.SuperCallExpr{BaseExpr: ast.NewBase(pos), Method: method, Args: args}
			}
			return &ast.GetFieldExpr{BaseExpr: ast.NewBase(pos), Object: &ast.EnlightsExpr{BaseExpr: ast.NewBase(pos)}, Field: method}
		}
		p.errorf("unexpected enlights without .")
		return &ast.EnlightsExpr{BaseExpr: ast.NewBase(pos)}
	case token.IDENT:
		name := p.cur.Lexeme
		p.advance()
		return &ast.IdentExpr{BaseExpr: ast.NewBase(pos), Name: name}
	case token.NEW:
		p.advance()
		className := p.cur.Lexeme
		p.expect(token.IDENT)
		p.expect(token.LPAREN)
		args := p.parseArgList()
		p.expect(token.RPAREN)
		return &ast.NewExpr{BaseExpr: ast.NewBase(pos), ClassName: className, Args: args}
	case token.LBRACK:
		p.advance()
		var elems []ast.Expr
		if p.cur.Type != token.RBRACK {
			for {
				elems = append(elems, p.parseExpression())
				if p.cur.Type != token.COMMA {
					break
				}
				p.advance()
			}
		}
		p.expect(token.RBRACK)
		return &ast.ArrayLitExpr{BaseExpr: ast.NewBase(pos), Elements: elems}
	case token.LPAREN:
		p.advance()
		inner := p.parseExpression()
		p.expect(token.RPAREN)
		return &ast.ParenExpr{BaseExpr: ast.NewBase(pos), Inner: inner}
	default:
		p.errorf("unexpected token in expression: %s", p.cur.Type)
		p.advance()
		return &ast.NullLit{BaseExpr: ast.NewBase(pos)}
	}
}