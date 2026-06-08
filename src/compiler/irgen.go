package compiler

import (
	"fmt"
	"strings"

	"github.com/unisc/compiladores/sol/src/ast"
	"github.com/unisc/compiladores/sol/src/semantic"
	"github.com/unisc/compiladores/sol/src/stdlib"
	"github.com/unisc/compiladores/sol/src/token"
)

// IRGen generates LLVM IR text from an annotated AST.
type IRGen struct {
	classes    map[string]*semantic.ClassInfo
	buf        strings.Builder
	globalDecl strings.Builder
	globals    map[string]string
	temp       int
	strID      int
}

func NewIRGen(classes map[string]*semantic.ClassInfo) *IRGen {
	return &IRGen{
		classes: classes,
		globals: make(map[string]string),
	}
}

func (g *IRGen) Generate(prog *ast.Program) string {
	var header strings.Builder
	header.WriteString("; SOL LLVM IR\n")
	header.WriteString("target triple = \"arm64-apple-macosx\"\n\n")
	header.WriteString("declare void @sol_panic(i8*)\n")
	header.WriteString("declare void @sol_print(i8*)\n")
	header.WriteString("declare i8* @sol_concat(i8*, i8*)\n")
	header.WriteString("declare i8* @sol_i64_to_str(i64)\n")
	header.WriteString("declare i8* @sol_f64_to_str(double)\n")
	header.WriteString("declare i8* @sol_bool_to_str(i1)\n")
	header.WriteString("declare %struct.SolObject* @sol_new(i8*)\n")
	header.WriteString("declare void @sol_set_field(%struct.SolObject*, i8*, i8*)\n")
	header.WriteString("declare i8* @sol_get_field(%struct.SolObject*, i8*)\n")
	g.writeRuntimeDecls(&header)
	header.WriteString("\n%struct.SolObject = type { i8*, i8*, i64, [64 x i8*], [64 x i8*] }\n\n")

	for _, decl := range prog.Decls {
		if c, ok := decl.(*ast.ClassDecl); ok {
			fmt.Fprintf(&header, "; struct %s\n", c.Name)
		}
	}

	g.buf.Reset()

	g.buf.WriteString("define i32 @main(i32 %argc, i8** %argv) {\n")
	g.buf.WriteString("entry:\n")
	g.buf.WriteString("  call void @sol_args_init(i32 %argc, i8** %argv)\n")
	for _, decl := range prog.Decls {
		if _, ok := decl.(*ast.ClassDecl); ok {
			continue
		}
		g.genTopLevel(decl)
	}
	g.buf.WriteString("  ret i32 0\n")
	g.buf.WriteString("}\n")

	for _, decl := range prog.Decls {
		if c, ok := decl.(*ast.ClassDecl); ok {
			for _, m := range c.Members {
				switch mm := m.(type) {
				case *ast.GlowDecl:
					g.genFunc(c.Name, "glow", mm.Body)
				case *ast.RayDecl:
					g.genFunc(c.Name, mm.Name, mm.Body)
				}
			}
		}
	}

	var out strings.Builder
	out.WriteString(header.String())
	out.WriteString(g.globalDecl.String())
	out.WriteString(g.buf.String())
	return out.String()
}

func (g *IRGen) genFunc(className, methodName string, body *ast.BlockStmt) {
	fmt.Fprintf(&g.buf, "define void @%s_%s() {\n", className, methodName)
	g.buf.WriteString("entry:\n")
	if body != nil {
		for _, s := range body.Stmts {
			g.genStmt(s)
		}
	}
	g.buf.WriteString("  ret void\n")
	g.buf.WriteString("}\n\n")
}

func (g *IRGen) genTopLevel(d ast.TopLevelDecl) {
	switch s := d.(type) {
	case *ast.VarDeclStmt:
		if s.Value != nil {
			val := g.genExpr(s.Value)
			g.globals[s.Name] = val
		}
	case *ast.ExprStmt:
		g.genExpr(s.Expr)
	case *ast.IfStmt:
		g.genIf(s)
	case *ast.WhileStmt:
		g.genWhile(s)
	case *ast.AssignStmt:
		val := g.genExpr(s.Value)
		if id, ok := s.Target.(*ast.IdentExpr); ok {
			g.globals[id.Name] = val
		}
	case *ast.ForEachStmt:
		g.genForEach(s)
	case *ast.TryCatchStmt:
		for _, st := range s.Try.Stmts {
			g.genStmt(st)
		}
	}
}

func (g *IRGen) genForEach(s *ast.ForEachStmt) {
	// Native backend: for-each is VM-first; emit body once without full iteration.
	for _, st := range s.Body.Stmts {
		g.genStmt(st)
	}
}

func (g *IRGen) genStmt(s ast.Stmt) {
	switch st := s.(type) {
	case *ast.VarDeclStmt:
		if st.Value != nil {
			g.genExpr(st.Value)
		}
	case *ast.AssignStmt:
		g.genExpr(st.Value)
	case *ast.IfStmt:
		g.genIf(st)
	case *ast.WhileStmt:
		g.genWhile(st)
	case *ast.ForEachStmt:
		g.genForEach(st)
	case *ast.EmitStmt:
		if st.Value != nil {
			g.genExpr(st.Value)
		}
	case *ast.FlareStmt:
		ptr := g.genStringValue(st.Value)
		fmt.Fprintf(&g.buf, "  call void @sol_panic(i8* %s)\n", ptr)
	case *ast.ExprStmt:
		g.genExpr(st.Expr)
	case *ast.BlockStmt:
		for _, inner := range st.Stmts {
			g.genStmt(inner)
		}
	case *ast.TryCatchStmt:
		for _, inner := range st.Try.Stmts {
			g.genStmt(inner)
		}
	}
}

func (g *IRGen) genIf(s *ast.IfStmt) {
	elseL := g.freshLabel()
	endL := g.freshLabel()
	cond := g.genExpr(s.Cond)
	fmt.Fprintf(&g.buf, "  br i1 %s, label %%%s, label %%%s\n", cond, elseL, endL)
	fmt.Fprintf(&g.buf, "%s:\n", elseL)
	for _, st := range s.Then.Stmts {
		g.genStmt(st)
	}
	if s.Else != nil {
		fmt.Fprintf(&g.buf, "  br label %%%s\n", endL)
		fmt.Fprintf(&g.buf, "%s:\n", endL)
		for _, st := range s.Else.Stmts {
			g.genStmt(st)
		}
	}
}

func (g *IRGen) genWhile(s *ast.WhileStmt) {
	start := g.freshLabel()
	end := g.freshLabel()
	fmt.Fprintf(&g.buf, "  br label %%%s\n", start)
	fmt.Fprintf(&g.buf, "%s:\n", start)
	cond := g.genExpr(s.Cond)
	fmt.Fprintf(&g.buf, "  br i1 %s, label %%%s_body, label %%%s\n", cond, start, end)
	fmt.Fprintf(&g.buf, "%s_body:\n", start)
	for _, st := range s.Body.Stmts {
		g.genStmt(st)
	}
	fmt.Fprintf(&g.buf, "  br label %%%s\n", start)
	fmt.Fprintf(&g.buf, "%s:\n", end)
}

func (g *IRGen) genExpr(e ast.Expr) string {
	switch ex := e.(type) {
	case *ast.IntLit:
		t := g.freshTemp()
		fmt.Fprintf(&g.buf, "  %s = add i64 0, %d\n", t, ex.Value)
		return t
	case *ast.FloatLit:
		t := g.freshTemp()
		fmt.Fprintf(&g.buf, "  %s = fadd double 0.0, %g\n", t, ex.Value)
		return t
	case *ast.StringLit:
		return g.stringConst(ex)
	case *ast.BoolLit:
		t := g.freshTemp()
		v := 0
		if ex.Value {
			v = 1
		}
		fmt.Fprintf(&g.buf, "  %s = icmp eq i1 %d, 1\n", t, v)
		return t
	case *ast.IdentExpr:
		if val, ok := g.globals[ex.Name]; ok {
			return val
		}
		t := g.freshTemp()
		fmt.Fprintf(&g.buf, "  %s = add i64 0, 0  ; load %s\n", t, ex.Name)
		return t
	case *ast.BinaryExpr:
		if ex.Operator == token.PLUS {
			lt := ex.Left.GetType()
			rt := ex.Right.GetType()
			if (lt != nil && lt.Base == "string") || (rt != nil && rt.Base == "string") {
				l := g.genStringValue(ex.Left)
				r := g.genStringValue(ex.Right)
				t := g.freshTemp()
				fmt.Fprintf(&g.buf, "  %s = call i8* @sol_concat(i8* %s, i8* %s)\n", t, l, r)
				return t
			}
		}
		l := g.genExpr(ex.Left)
		r := g.genExpr(ex.Right)
		t := g.freshTemp()
		op := irOp(ex.Operator)
		fmt.Fprintf(&g.buf, "  %s = %s i64 %s, %s\n", t, op, l, r)
		return t
	case *ast.CallExpr:
		if gf, ok := ex.Callee.(*ast.GetFieldExpr); ok {
			if id, ok := gf.Object.(*ast.IdentExpr); ok && stdlib.IsBuiltin(id.Name) {
				if _, ok := stdlib.Lookup(id.Name, gf.Field); ok {
					return g.genBuiltinCall(id.Name, gf.Field, ex.Args)
				}
			}
			for _, arg := range ex.Args {
				g.genExpr(arg)
			}
			fmt.Fprintf(&g.buf, "  call void @%s_%s()\n", g.exprClass(gf.Object), gf.Field)
		}
		return g.freshTemp()
	case *ast.NewExpr:
		t := g.freshTemp()
		classPtr := g.stringConst(&ast.StringLit{Value: ex.ClassName})
		fmt.Fprintf(&g.buf, "  %s = call %%struct.SolObject* @sol_new(i8* %s)\n", t, classPtr)
		fmt.Fprintf(&g.buf, "  call void @%s_glow()\n", ex.ClassName)
		return t
	case *ast.GetFieldExpr:
		if ex.Field == "length" {
			t := g.freshTemp()
			fmt.Fprintf(&g.buf, "  %s = add i64 0, 0  ; length placeholder\n", t)
			return t
		}
		return g.freshTemp()
	case *ast.ParenExpr:
		return g.genExpr(ex.Inner)
	default:
		return g.freshTemp()
	}
}

func (g *IRGen) genConsolePrint(args []ast.Expr) {
	if len(args) == 0 {
		return
	}
	line := g.genStringValue(args[0])
	for i := 1; i < len(args); i++ {
		space := g.stringConst(&ast.StringLit{Value: " "})
		combined := g.freshTemp()
		fmt.Fprintf(&g.buf, "  %s = call i8* @sol_concat(i8* %s, i8* %s)\n", combined, line, space)
		next := g.genStringValue(args[i])
		line = g.freshTemp()
		fmt.Fprintf(&g.buf, "  %s = call i8* @sol_concat(i8* %s, i8* %s)\n", line, combined, next)
	}
	fmt.Fprintf(&g.buf, "  call void @sol_print(i8* %s)\n", line)
}

func (g *IRGen) genStringValue(e ast.Expr) string {
	switch ex := e.(type) {
	case *ast.StringLit:
		return g.stringConst(ex)
	case *ast.IntLit:
		num := g.freshTemp()
		fmt.Fprintf(&g.buf, "  %s = add i64 0, %d\n", num, ex.Value)
		str := g.freshTemp()
		fmt.Fprintf(&g.buf, "  %s = call i8* @sol_i64_to_str(i64 %s)\n", str, num)
		return str
	case *ast.FloatLit:
		num := g.freshTemp()
		fmt.Fprintf(&g.buf, "  %s = fadd double 0.0, %g\n", num, ex.Value)
		str := g.freshTemp()
		fmt.Fprintf(&g.buf, "  %s = call i8* @sol_f64_to_str(double %s)\n", str, num)
		return str
	case *ast.BoolLit:
		v := 0
		if ex.Value {
			v = 1
		}
		str := g.freshTemp()
		fmt.Fprintf(&g.buf, "  %s = call i8* @sol_bool_to_str(i1 %d)\n", str, v)
		return str
	case *ast.BinaryExpr:
		if ex.Operator == token.PLUS {
			l := g.genStringValue(ex.Left)
			r := g.genStringValue(ex.Right)
			t := g.freshTemp()
			fmt.Fprintf(&g.buf, "  %s = call i8* @sol_concat(i8* %s, i8* %s)\n", t, l, r)
			return t
		}
	}
	return g.stringConst(&ast.StringLit{Value: "?"})
}

func (g *IRGen) stringConst(e ast.Expr) string {
	s, ok := e.(*ast.StringLit)
	if !ok {
		return "null"
	}
	name := fmt.Sprintf("@.str.%d", g.strID)
	g.strID++
	fmt.Fprintf(&g.globalDecl, "%s = private unnamed_addr constant [%d x i8] c\"%s\\00\"\n", name, len(s.Value)+1, s.Value)
	t := g.freshTemp()
	fmt.Fprintf(&g.buf, "  %s = getelementptr [%d x i8], [%d x i8]* %s, i64 0, i64 0\n", t, len(s.Value)+1, len(s.Value)+1, name)
	return t
}

func (g *IRGen) exprClass(e ast.Expr) string {
	if t := e.GetType(); t != nil {
		return t.Base
	}
	return "Object"
}

func (g *IRGen) freshTemp() string {
	name := fmt.Sprintf("%%t%d", g.temp)
	g.temp++
	return name
}

func (g *IRGen) freshLabel() string {
	name := fmt.Sprintf("L%d", g.temp)
	g.temp++
	return name
}

func irOp(op token.Type) string {
	switch op {
	case token.PLUS:
		return "add"
	case token.MINUS:
		return "sub"
	case token.ASTERISK:
		return "mul"
	case token.EQ:
		return "icmp eq"
	case token.NOT_EQ:
		return "icmp ne"
	case token.LT:
		return "icmp slt"
	case token.GT:
		return "icmp sgt"
	case token.LT_EQ:
		return "icmp sle"
	case token.GT_EQ:
		return "icmp sge"
	default:
		return "add"
	}
}
