package parser

import (
	"strconv"

	plexer "github.com/alecthomas/participle/v2/lexer"

	"github.com/unisc/compiladores/sol/src/ast"
	"github.com/unisc/compiladores/sol/src/token"
)

func pos(p plexer.Position) ast.Pos {
	return ast.Pos{Line: p.Line, Column: p.Column}
}

// ── Program / top-level ───────────────────────────────────────────────────────

func convertProgram(p *PProgram) *ast.Program {
	prog := &ast.Program{}
	for _, d := range p.Decls {
		prog.Decls = append(prog.Decls, convertTopDecl(d)...)
	}
	return prog
}

func convertTopDecl(d *PTopDecl) []ast.TopLevelDecl {
	switch {
	case d.Class != nil:
		return []ast.TopLevelDecl{convertClassDecl(d.Class)}
	case d.Import != nil:
		return []ast.TopLevelDecl{convertImport(d.Import)}
	case d.Alias != nil:
		return []ast.TopLevelDecl{convertAlias(d.Alias)}
	default:
		return []ast.TopLevelDecl{convertStmt(d.Stmt).(ast.TopLevelDecl)}
	}
}

// ── Classes ───────────────────────────────────────────────────────────────────

func convertClassDecl(c *PClassDecl) *ast.ClassDecl {
	decl := &ast.ClassDecl{PosInfo: pos(c.Pos), Name: c.Name, SuperName: c.Super}
	for _, m := range c.Members {
		decl.Members = append(decl.Members, convertMember(m))
	}
	return decl
}

func convertMember(m *PMember) ast.MemberDecl {
	if m.GlowDecl != nil {
		return convertGlowDecl(m.GlowDecl)
	}
	return convertVisMember(m.WithVis)
}

func convertVisMember(v *PVisMember) ast.MemberDecl {
	if v.Ray != nil {
		return convertRayDecl(v.Vis, v.Ray)
	}
	return convertFieldDecl(v.Vis, v.Field)
}

func convertGlowDecl(g *PGlowDecl) *ast.GlowDecl {
	return &ast.GlowDecl{
		PosInfo: pos(g.Pos),
		Params:  convertParams(g.Params),
		Body:    convertBlock(g.Body),
	}
}

func convertRayDecl(vis string, r *PRayBody) *ast.RayDecl {
	return &ast.RayDecl{
		PosInfo:    pos(r.Pos),
		Visibility: vis,
		Name:       r.Name,
		Params:     convertParams(r.Params),
		ReturnType: convertTypeOpt(r.RetType),
		Body:       convertBlock(r.Body),
	}
}

func convertFieldDecl(vis string, f *PFieldBody) *ast.FieldDecl {
	return &ast.FieldDecl{
		PosInfo:    pos(f.Pos),
		Visibility: vis,
		Type:       convertType(f.Type),
		Name:       f.Name,
	}
}

func convertParams(ps []*PParam) []ast.Param {
	var out []ast.Param
	for _, p := range ps {
		out = append(out, ast.Param{
			Type: convertType(p.Type),
			Name: p.Name,
			Pos:  pos(p.Pos),
		})
	}
	return out
}

// ── Types ─────────────────────────────────────────────────────────────────────

func convertType(t *PType) *ast.TypeDesc {
	base := &ast.TypeDesc{Pos: pos(t.Pos), Base: t.Base}
	result := base
	for range t.Arrays {
		result = &ast.TypeDesc{Pos: pos(t.Pos), IsArray: true, ElemType: result}
	}
	return result
}

func convertTypeOpt(t *PType) *ast.TypeDesc {
	if t == nil {
		return nil
	}
	return convertType(t)
}

// ── Import / Alias ────────────────────────────────────────────────────────────

func convertImport(i *PImport) *ast.ImportDecl {
	return &ast.ImportDecl{PosInfo: pos(i.Pos), Path: i.Path}
}

func convertAlias(a *PAlias) *ast.TypeAliasDecl {
	return &ast.TypeAliasDecl{PosInfo: pos(a.Pos), Name: a.Name, Type: convertType(a.Type)}
}

// ── Statements ────────────────────────────────────────────────────────────────

func convertBlock(b *PBlock) *ast.BlockStmt {
	block := &ast.BlockStmt{PosInfo: pos(b.Pos)}
	for _, s := range b.Stmts {
		block.Stmts = append(block.Stmts, convertStmt(s))
	}
	return block
}

func convertStmt(s *PStmt) ast.Stmt {
	switch {
	case s.Var != nil:
		return convertVarDecl(s.Var)
	case s.If != nil:
		return convertIfStmt(s.If)
	case s.While != nil:
		return convertWhileStmt(s.While)
	case s.For != nil:
		return convertForStmt(s.For)
	case s.Break != nil:
		return &ast.BreakStmt{PosInfo: pos(s.Break.Pos)}
	case s.Continue != nil:
		return &ast.ContinueStmt{PosInfo: pos(s.Continue.Pos)}
	case s.Emit != nil:
		return convertEmitStmt(s.Emit)
	case s.Flare != nil:
		return convertFlareStmt(s.Flare)
	case s.Try != nil:
		return convertTryCatch(s.Try)
	case s.Block != nil:
		return convertBlock(s.Block)
	default:
		return convertAssignOrExpr(s.Expr)
	}
}

func convertVarDecl(v *PVarDecl) *ast.VarDeclStmt {
	return &ast.VarDeclStmt{
		PosInfo: pos(v.Pos),
		Name:    v.Name,
		Type:    convertType(v.Type),
		Value:   convertExprOpt(v.Value),
	}
}

func convertIfStmt(i *PIfStmt) *ast.IfStmt {
	var els *ast.BlockStmt
	if i.Else != nil {
		els = convertBlock(i.Else)
	}
	return &ast.IfStmt{
		PosInfo: pos(i.Pos),
		Cond:    convertExpr(i.Cond),
		Then:    convertBlock(i.Then),
		Else:    els,
	}
}

func convertWhileStmt(w *PWhileStmt) *ast.WhileStmt {
	return &ast.WhileStmt{
		PosInfo: pos(w.Pos),
		Cond:    convertExpr(w.Cond),
		Body:    convertBlock(w.Body),
	}
}

func convertForStmt(f *PForStmt) ast.Stmt {
	if f.Each != nil {
		return &ast.ForEachStmt{
			PosInfo: pos(f.Each.Pos),
			VarName: f.Each.Var,
			Iter:    convertExpr(f.Each.Iter),
			Body:    convertBlock(f.Each.Body),
		}
	}
	r := f.Range
	return &ast.ForRangeStmt{
		PosInfo: pos(r.Pos),
		VarName: r.Var,
		Start:   convertExpr(r.Start),
		End:     convertExpr(r.End),
		Body:    convertBlock(r.Body),
	}
}

func convertEmitStmt(e *PEmitStmt) *ast.EmitStmt {
	return &ast.EmitStmt{PosInfo: pos(e.Pos), Value: convertExprOpt(e.Value)}
}

func convertFlareStmt(f *PFlareStmt) *ast.FlareStmt {
	return &ast.FlareStmt{PosInfo: pos(f.Pos), Value: convertExpr(f.Value)}
}

func convertTryCatch(t *PTryCatch) *ast.TryCatchStmt {
	return &ast.TryCatchStmt{
		PosInfo:  pos(t.Pos),
		Try:      convertBlock(t.Try),
		CatchVar: t.CatchVar,
		Catch:    convertBlock(t.Catch),
	}
}

func convertAssignOrExpr(a *PAssignOrExpr) ast.Stmt {
	left := convertExpr(a.Left)
	if a.Right != nil {
		return &ast.AssignStmt{PosInfo: pos(a.Pos), Target: left, Value: convertExpr(a.Right)}
	}
	return &ast.ExprStmt{PosInfo: pos(a.Pos), Expr: left}
}

// ── Expressions ───────────────────────────────────────────────────────────────

func convertExprOpt(e *PExpr) ast.Expr {
	if e == nil {
		return nil
	}
	return convertExpr(e)
}

func convertExpr(e *PExpr) ast.Expr {
	result := convertAndExpr(e.Left)
	for _, rhs := range e.Rest {
		right := convertAndExpr(rhs.Right)
		result = &ast.BinaryExpr{
			BaseExpr: ast.NewBase(result.Pos()),
			Left:     result,
			Operator: token.OR,
			Right:    right,
		}
	}
	return result
}

func convertAndExpr(e *PAndExpr) ast.Expr {
	result := convertEqExpr(e.Left)
	for _, rhs := range e.Rest {
		right := convertEqExpr(rhs.Right)
		result = &ast.BinaryExpr{
			BaseExpr: ast.NewBase(result.Pos()),
			Left:     result,
			Operator: token.AND,
			Right:    right,
		}
	}
	return result
}

func convertEqExpr(e *PEqExpr) ast.Expr {
	result := convertRelExpr(e.Left)
	for _, rhs := range e.Rest {
		right := convertRelExpr(rhs.Right)
		result = &ast.BinaryExpr{
			BaseExpr: ast.NewBase(result.Pos()),
			Left:     result,
			Operator: opFromStr(rhs.Op),
			Right:    right,
		}
	}
	return result
}

func convertRelExpr(e *PRelExpr) ast.Expr {
	result := convertAddExpr(e.Left)
	for _, rhs := range e.Rest {
		right := convertAddExpr(rhs.Right)
		result = &ast.BinaryExpr{
			BaseExpr: ast.NewBase(result.Pos()),
			Left:     result,
			Operator: opFromStr(rhs.Op),
			Right:    right,
		}
	}
	return result
}

func convertAddExpr(e *PAddExpr) ast.Expr {
	result := convertMulExpr(e.Left)
	for _, rhs := range e.Rest {
		right := convertMulExpr(rhs.Right)
		result = &ast.BinaryExpr{
			BaseExpr: ast.NewBase(result.Pos()),
			Left:     result,
			Operator: opFromStr(rhs.Op),
			Right:    right,
		}
	}
	return result
}

func convertMulExpr(e *PMulExpr) ast.Expr {
	result := convertUnary(e.Left)
	for _, rhs := range e.Rest {
		right := convertUnary(rhs.Right)
		result = &ast.BinaryExpr{
			BaseExpr: ast.NewBase(result.Pos()),
			Left:     result,
			Operator: opFromStr(rhs.Op),
			Right:    right,
		}
	}
	return result
}

func convertUnary(u *PUnary) ast.Expr {
	if u.Op != nil {
		return &ast.UnaryExpr{
			BaseExpr: ast.NewBase(pos(u.Op.Pos)),
			Operator: opFromStr(u.Op.Sym),
			Operand:  convertUnary(u.Op.Operand),
		}
	}
	return convertPostfix(u.Base)
}

func convertPostfix(p *PPostfix) ast.Expr {
	result := convertPrimary(p.Base)
	for _, op := range p.Ops {
		switch {
		case op.Field != nil:
			field := op.Field
			gf := &ast.GetFieldExpr{
				BaseExpr: ast.NewBase(result.Pos()),
				Object:   result,
				Field:    field.Name,
			}
			if field.Call != nil {
				result = &ast.CallExpr{
					BaseExpr: ast.NewBase(result.Pos()),
					Callee:   gf,
					Args:     convertArgs(field.Call.Args),
				}
			} else {
				result = gf
			}
		case op.Index != nil:
			result = &ast.IndexExpr{
				BaseExpr: ast.NewBase(result.Pos()),
				Object:   result,
				Index:    convertExpr(op.Index.Index),
			}
		case op.Call != nil:
			result = &ast.CallExpr{
				BaseExpr: ast.NewBase(result.Pos()),
				Callee:   result,
				Args:     convertArgs(op.Call.Args),
			}
		}
	}
	return result
}

func convertPrimary(p *PPrimary) ast.Expr {
	base := ast.NewBase(pos(p.Pos))
	switch {
	case p.Int != "":
		v, _ := strconv.ParseInt(p.Int, 10, 64)
		return &ast.IntLit{BaseExpr: base, Value: v}
	case p.Float != "":
		v, _ := strconv.ParseFloat(p.Float, 64)
		return &ast.FloatLit{BaseExpr: base, Value: v}
	case p.Str != "":
		return &ast.StringLit{BaseExpr: base, Value: p.Str}
	case p.Bool != "":
		return &ast.BoolLit{BaseExpr: base, Value: p.Bool == "true"}
	case p.Null != "":
		return &ast.NullLit{BaseExpr: base}
	case p.This != "":
		return &ast.ThisExpr{BaseExpr: base}
	case p.Super != nil:
		s := p.Super
		if s.Call != nil {
			return &ast.SuperCallExpr{
				BaseExpr: ast.NewBase(pos(s.Pos)),
				Method:   s.Method,
				Args:     convertArgs(s.Call.Args),
			}
		}
		return &ast.GetFieldExpr{
			BaseExpr: ast.NewBase(pos(s.Pos)),
			Object:   &ast.RadiateExpr{BaseExpr: ast.NewBase(pos(s.Pos))},
			Field:    s.Method,
		}
	case p.New != nil:
		n := p.New
		return &ast.NewExpr{
			BaseExpr:  ast.NewBase(pos(n.Pos)),
			ClassName: n.Class,
			Args:      convertArgs(n.Args),
		}
	case p.Array != nil:
		a := p.Array
		arr := &ast.ArrayLitExpr{BaseExpr: ast.NewBase(pos(a.Pos))}
		for _, item := range a.Items {
			arr.Elements = append(arr.Elements, convertExpr(item))
		}
		return arr
	case p.Paren != nil:
		return &ast.ParenExpr{BaseExpr: base, Inner: convertExpr(p.Paren.Inner)}
	default:
		return &ast.IdentExpr{BaseExpr: base, Name: p.Ident}
	}
}

func convertArgs(args []*PExpr) []ast.Expr {
	var out []ast.Expr
	for _, a := range args {
		out = append(out, convertExpr(a))
	}
	return out
}

// opFromStr maps a lexeme string to the corresponding token.Type operator.
func opFromStr(op string) token.Type {
	switch op {
	case "+":
		return token.PLUS
	case "-":
		return token.MINUS
	case "*":
		return token.ASTERISK
	case "/":
		return token.SLASH
	case "%":
		return token.PERCENT
	case "!":
		return token.BANG
	case "==":
		return token.EQ
	case "!=":
		return token.NOT_EQ
	case "<":
		return token.LT
	case ">":
		return token.GT
	case "<=":
		return token.LT_EQ
	case ">=":
		return token.GT_EQ
	case "&&":
		return token.AND
	case "||":
		return token.OR
	}
	return token.ILLEGAL
}
