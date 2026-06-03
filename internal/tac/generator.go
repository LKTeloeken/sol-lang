package tac

import (
	"fmt"
	"strings"

	"github.com/unisc/compiladores/sol/internal/ast"
	"github.com/unisc/compiladores/sol/internal/semantic"
	"github.com/unisc/compiladores/sol/internal/token"
)

type Op string

const (
	OpAssign   Op = "assign"
	OpBinOp    Op = "binop"
	OpLabel    Op = "label"
	OpGoto     Op = "goto"
	OpIfFalse  Op = "ifFalse"
	OpParam    Op = "param"
	OpCall     Op = "call"
	OpThrow    Op = "throw"
	OpNew      Op = "new"
	OpReturn   Op = "return"
	OpComment  Op = "comment"
	OpBeginTry Op = "beginTry"
	OpEndTry   Op = "endTry"
	OpPrint    Op = "print"
	OpReadLine Op = "readLine"
	OpReadInt  Op = "readInt"
	OpFileRead Op = "fileRead"
	OpFileWrite Op = "fileWrite"
	OpArrayLit Op = "arrayLit"
)

const BuiltinConsole = "Console"
const BuiltinFile = "File"

type loopContext struct {
	breakLabel    string
	continueLabel string
}

type Instr struct {
	Op       Op
	Result   string
	Arg1     string
	Arg2     string
	Operator string
	Label    string
	Comment  string
}

func (i Instr) String() string {
	switch i.Op {
	case OpAssign:
		if i.Comment != "" {
			return fmt.Sprintf("%s = %s  ; %s", i.Result, i.Arg1, i.Comment)
		}
		return fmt.Sprintf("%s = %s", i.Result, i.Arg1)
	case OpBinOp:
		return fmt.Sprintf("%s = %s %s %s", i.Result, i.Arg1, i.Operator, i.Arg2)
	case OpLabel:
		return fmt.Sprintf("%s:", i.Label)
	case OpGoto:
		return fmt.Sprintf("goto %s", i.Label)
	case OpIfFalse:
		return fmt.Sprintf("ifFalse %s goto %s", i.Arg1, i.Label)
	case OpParam:
		return fmt.Sprintf("param %s", i.Arg1)
	case OpCall:
		return fmt.Sprintf("call %s, %s", i.Arg1, i.Arg2)
	case OpThrow:
		return fmt.Sprintf("throw %s", i.Arg1)
	case OpNew:
		return fmt.Sprintf("%s = new %s", i.Result, i.Arg1)
	case OpReturn:
		if i.Arg1 != "" {
			return fmt.Sprintf("return %s", i.Arg1)
		}
		return "return"
	case OpComment:
		return "; " + i.Comment
	case OpBeginTry:
		return fmt.Sprintf("beginTry %s %s", i.Label, i.Arg1)
	case OpEndTry:
		return fmt.Sprintf("endTry %s", i.Label)
	case OpPrint:
		return fmt.Sprintf("print %s", i.Arg2)
	case OpReadLine:
		if i.Result != "" {
			return fmt.Sprintf("%s = readLine %s", i.Result, i.Arg2)
		}
		return fmt.Sprintf("readLine %s", i.Arg2)
	case OpReadInt:
		if i.Result != "" {
			return fmt.Sprintf("%s = readInt", i.Result)
		}
		return "readInt"
	case OpFileRead:
		return fmt.Sprintf("%s = fileRead %s", i.Result, i.Arg1)
	case OpFileWrite:
		return fmt.Sprintf("fileWrite %s", i.Arg1)
	case OpArrayLit:
		return fmt.Sprintf("%s = arrayLit %s", i.Result, i.Arg2)
	default:
		return fmt.Sprintf("; unknown op %s", i.Op)
	}
}

type Generator struct {
	instrs    []Instr
	temp      int
	label     int
	classes   map[string]*semantic.ClassInfo
	curClass  string
	loopStack []loopContext
}

func New(classes map[string]*semantic.ClassInfo) *Generator {
	return &Generator{classes: classes}
}

func (g *Generator) Generate(prog *ast.Program) string {
	g.Build(prog)
	var b strings.Builder
	for _, ins := range g.instrs {
		b.WriteString(ins.String())
		b.WriteByte('\n')
	}
	return b.String()
}

// Build emits TAC instructions for prog.
func (g *Generator) Build(prog *ast.Program) {
	g.instrs = nil
	g.temp = 0
	g.label = 0
	for _, decl := range prog.Decls {
		if c, ok := decl.(*ast.ClassDecl); ok {
			g.genClass(c)
		}
	}
	g.emit(Instr{Op: OpLabel, Label: "__program", Comment: "program entry"})
	for _, decl := range prog.Decls {
		if _, ok := decl.(*ast.ClassDecl); ok {
			continue
		}
		g.genTopLevel(decl)
	}
	g.emit(Instr{Op: OpLabel, Label: "__end"})
}

// Instructions returns the emitted TAC instruction slice.
func (g *Generator) Instructions() []Instr {
	return g.instrs
}

func (g *Generator) genClass(c *ast.ClassDecl) {
	g.curClass = c.Name
	g.emit(Instr{Op: OpComment, Comment: fmt.Sprintf("class %s", c.Name)})
	for _, m := range c.Members {
		switch mm := m.(type) {
		case *ast.GlowDecl:
			g.genMethod(c.Name, "glow", mm.Body, nil)
		case *ast.RayDecl:
			g.genMethod(c.Name, mm.Name, mm.Body, mm.ReturnType)
		}
	}
	g.curClass = ""
}

func (g *Generator) genMethod(className, methodName string, body *ast.BlockStmt, retType *ast.TypeDesc) {
	g.emit(Instr{Op: OpLabel, Label: fmt.Sprintf("%s.%s", className, methodName)})
	if body != nil {
		for _, s := range body.Stmts {
			g.genStmt(s)
		}
	}
	if retType == nil || retType.Base == "void" {
		g.emit(Instr{Op: OpReturn})
	}
}

func (g *Generator) genTopLevel(d ast.TopLevelDecl) {
	switch s := d.(type) {
	case *ast.VarDeclStmt:
		g.genVarDecl(s)
	case *ast.AssignStmt:
		g.genAssign(s)
	case *ast.IfStmt:
		g.genIf(s)
	case *ast.WhileStmt:
		g.genWhile(s)
	case *ast.ForEachStmt:
		g.genForEach(s)
	case *ast.ForRangeStmt:
		g.genForRange(s)
	case *ast.FlareStmt:
		g.genFlare(s)
	case *ast.TryCatchStmt:
		g.genTryCatch(s)
	case *ast.ExprStmt:
		g.genExpr(s.Expr)
	}
}

func (g *Generator) genStmt(s ast.Stmt) {
	switch st := s.(type) {
	case *ast.VarDeclStmt:
		g.genVarDecl(st)
	case *ast.AssignStmt:
		g.genAssign(st)
	case *ast.IfStmt:
		g.genIf(st)
	case *ast.WhileStmt:
		g.genWhile(st)
	case *ast.ForEachStmt:
		g.genForEach(st)
	case *ast.ForRangeStmt:
		g.genForRange(st)
	case *ast.EmitStmt:
		if st.Value != nil {
			v := g.genExpr(st.Value)
			g.emit(Instr{Op: OpReturn, Arg1: v})
		} else {
			g.emit(Instr{Op: OpReturn})
		}
	case *ast.FlareStmt:
		g.genFlare(st)
	case *ast.TryCatchStmt:
		g.genTryCatch(st)
	case *ast.ExprStmt:
		g.genExpr(st.Expr)
	case *ast.BreakStmt:
		g.genBreak()
	case *ast.ContinueStmt:
		g.genContinue()
	case *ast.BlockStmt:
		for _, inner := range st.Stmts {
			g.genStmt(inner)
		}
	}
}

func (g *Generator) genVarDecl(v *ast.VarDeclStmt) {
	if v.Value != nil {
		val := g.genExpr(v.Value)
		g.emit(Instr{Op: OpAssign, Result: v.Name, Arg1: val})
	}
}

func (g *Generator) genAssign(s *ast.AssignStmt) {
	val := g.genExpr(s.Value)
	target := g.genLValue(s.Target)
	g.emit(Instr{Op: OpAssign, Result: target, Arg1: val})
}

func (g *Generator) genIf(s *ast.IfStmt) {
	cond := g.genExpr(s.Cond)
	elseLabel := g.freshLabel()
	endLabel := g.freshLabel()
	if s.Else == nil {
		endLabel = elseLabel
	}
	g.emit(Instr{Op: OpIfFalse, Arg1: cond, Label: elseLabel})
	for _, st := range s.Then.Stmts {
		g.genStmt(st)
	}
	if s.Else != nil {
		g.emit(Instr{Op: OpGoto, Label: endLabel})
		g.emit(Instr{Op: OpLabel, Label: elseLabel})
		for _, st := range s.Else.Stmts {
			g.genStmt(st)
		}
		g.emit(Instr{Op: OpLabel, Label: endLabel})
	} else {
		g.emit(Instr{Op: OpLabel, Label: elseLabel})
	}
}

func (g *Generator) pushLoop(breakLabel, continueLabel string) {
	g.loopStack = append(g.loopStack, loopContext{breakLabel: breakLabel, continueLabel: continueLabel})
}

func (g *Generator) popLoop() {
	g.loopStack = g.loopStack[:len(g.loopStack)-1]
}

func (g *Generator) genBreak() {
	if len(g.loopStack) == 0 {
		return
	}
	g.emit(Instr{Op: OpGoto, Label: g.loopStack[len(g.loopStack)-1].breakLabel})
}

func (g *Generator) genContinue() {
	if len(g.loopStack) == 0 {
		return
	}
	g.emit(Instr{Op: OpGoto, Label: g.loopStack[len(g.loopStack)-1].continueLabel})
}

func (g *Generator) genWhile(s *ast.WhileStmt) {
	start := g.freshLabel()
	end := g.freshLabel()
	g.pushLoop(end, start)
	g.emit(Instr{Op: OpLabel, Label: start})
	cond := g.genExpr(s.Cond)
	g.emit(Instr{Op: OpIfFalse, Arg1: cond, Label: end})
	for _, st := range s.Body.Stmts {
		g.genStmt(st)
	}
	g.emit(Instr{Op: OpGoto, Label: start})
	g.emit(Instr{Op: OpLabel, Label: end})
	g.popLoop()
}

func (g *Generator) genForEach(s *ast.ForEachStmt) {
	arr := g.genExpr(s.Iter)
	idx := g.freshTemp()
	lenTmp := g.freshTemp()
	g.emit(Instr{Op: OpAssign, Result: idx, Arg1: "0", Comment: "for each index"})
	loopStart := g.freshLabel()
	loopEnd := g.freshLabel()
	contLabel := g.freshLabel()
	g.pushLoop(loopEnd, contLabel)
	g.emit(Instr{Op: OpLabel, Label: loopStart})
	g.emit(Instr{Op: OpAssign, Result: lenTmp, Arg1: arr + ".length", Comment: "array length"})
	lt := g.freshTemp()
	g.emit(Instr{Op: OpBinOp, Result: lt, Arg1: idx, Operator: "<", Arg2: lenTmp})
	g.emit(Instr{Op: OpIfFalse, Arg1: lt, Label: loopEnd})
	g.emit(Instr{Op: OpAssign, Result: s.VarName, Arg1: arr + "[" + idx + "]", Comment: "for each elem"})
	for _, st := range s.Body.Stmts {
		g.genStmt(st)
	}
	g.emit(Instr{Op: OpLabel, Label: contLabel})
	inc := g.freshTemp()
	g.emit(Instr{Op: OpBinOp, Result: inc, Arg1: idx, Operator: "+", Arg2: "1"})
	g.emit(Instr{Op: OpAssign, Result: idx, Arg1: inc})
	g.emit(Instr{Op: OpGoto, Label: loopStart})
	g.emit(Instr{Op: OpLabel, Label: loopEnd, Comment: "end for each " + arr})
	g.popLoop()
}

func (g *Generator) genForRange(s *ast.ForRangeStmt) {
	startVal := g.genExpr(s.Start)
	endVal := g.genExpr(s.End)
	g.emit(Instr{Op: OpAssign, Result: s.VarName, Arg1: startVal, Comment: "for range init"})
	loopStart := g.freshLabel()
	loopEnd := g.freshLabel()
	contLabel := g.freshLabel()
	g.pushLoop(loopEnd, contLabel)
	g.emit(Instr{Op: OpLabel, Label: loopStart})
	lt := g.freshTemp()
	g.emit(Instr{Op: OpBinOp, Result: lt, Arg1: s.VarName, Operator: "<", Arg2: endVal})
	g.emit(Instr{Op: OpIfFalse, Arg1: lt, Label: loopEnd})
	for _, st := range s.Body.Stmts {
		g.genStmt(st)
	}
	g.emit(Instr{Op: OpLabel, Label: contLabel})
	inc := g.freshTemp()
	g.emit(Instr{Op: OpBinOp, Result: inc, Arg1: s.VarName, Operator: "+", Arg2: "1"})
	g.emit(Instr{Op: OpAssign, Result: s.VarName, Arg1: inc})
	g.emit(Instr{Op: OpGoto, Label: loopStart})
	g.emit(Instr{Op: OpLabel, Label: loopEnd, Comment: "end for range"})
	g.popLoop()
}

func (g *Generator) genFlare(s *ast.FlareStmt) {
	val := g.genExpr(s.Value)
	g.emit(Instr{Op: OpThrow, Arg1: val})
}

func (g *Generator) genTryCatch(s *ast.TryCatchStmt) {
	catchLabel := g.freshLabel()
	endLabel := g.freshLabel()
	g.emit(Instr{Op: OpBeginTry, Label: catchLabel, Arg1: s.CatchVar})
	for _, st := range s.Try.Stmts {
		g.genStmt(st)
	}
	g.emit(Instr{Op: OpEndTry, Label: endLabel})
	g.emit(Instr{Op: OpLabel, Label: catchLabel, Comment: "catch " + s.CatchVar})
	for _, st := range s.Catch.Stmts {
		g.genStmt(st)
	}
	g.emit(Instr{Op: OpLabel, Label: endLabel})
}

func (g *Generator) genExpr(e ast.Expr) string {
	switch ex := e.(type) {
	case *ast.IntLit:
		return fmt.Sprintf("%d", ex.Value)
	case *ast.FloatLit:
		return fmt.Sprintf("%g", ex.Value)
	case *ast.StringLit:
		return fmt.Sprintf("%q", ex.Value)
	case *ast.BoolLit:
		if ex.Value {
			return "true"
		}
		return "false"
	case *ast.NullLit:
		return "null"
	case *ast.IdentExpr:
		return ex.Name
	case *ast.ThisExpr:
		return "this"
	case *ast.EnlightsExpr:
		return "enlights"
	case *ast.BinaryExpr:
		t := g.freshTemp()
		op := tokenOp(ex.Operator)
		g.emit(Instr{Op: OpBinOp, Result: t, Arg1: g.genExpr(ex.Left), Operator: op, Arg2: g.genExpr(ex.Right)})
		return t
	case *ast.UnaryExpr:
		t := g.freshTemp()
		if ex.Operator == token.BANG {
			g.emit(Instr{Op: OpBinOp, Result: t, Arg1: "!", Operator: "", Arg2: g.genExpr(ex.Operand)})
		} else {
			g.emit(Instr{Op: OpBinOp, Result: t, Arg1: "-", Operator: "", Arg2: g.genExpr(ex.Operand)})
		}
		return t
	case *ast.CallExpr:
		return g.genCall(ex)
	case *ast.NewExpr:
		t := g.freshTemp()
		g.emit(Instr{Op: OpNew, Result: t, Arg1: ex.ClassName})
		g.emit(Instr{Op: OpParam, Arg1: t})
		for _, arg := range ex.Args {
			g.emit(Instr{Op: OpParam, Arg1: g.genExpr(arg)})
		}
		g.emit(Instr{Op: OpCall, Arg1: ex.ClassName + ".glow", Arg2: fmt.Sprintf("%d", 1+len(ex.Args))})
		return t
	case *ast.GetFieldExpr:
		return g.genField(ex)
	case *ast.SuperCallExpr:
		g.emit(Instr{Op: OpParam, Arg1: "this"})
		for _, arg := range ex.Args {
			g.emit(Instr{Op: OpParam, Arg1: g.genExpr(arg)})
		}
		superClass := ""
		if g.curClass != "" {
			if ci, ok := g.classes[g.curClass]; ok && ci.Super != nil {
				superClass = ci.SuperName
			}
		}
		g.emit(Instr{Op: OpCall, Arg1: superClass + "." + ex.Method, Arg2: fmt.Sprintf("%d", 1+len(ex.Args))})
		t := g.freshTemp()
		g.emit(Instr{Op: OpAssign, Result: t, Arg1: "call_result", Comment: "super " + ex.Method})
		return t
	case *ast.ParenExpr:
		return g.genExpr(ex.Inner)
	case *ast.ArrayLitExpr:
		t := g.freshTemp()
		for _, el := range ex.Elements {
			g.emit(Instr{Op: OpParam, Arg1: g.genExpr(el)})
		}
		g.emit(Instr{Op: OpArrayLit, Result: t, Arg2: fmt.Sprintf("%d", len(ex.Elements))})
		return t
	case *ast.IndexExpr:
		obj := g.genExpr(ex.Object)
		idx := g.genExpr(ex.Index)
		return obj + "[" + idx + "]"
	default:
		return g.freshTemp()
	}
}

func (g *Generator) genCall(c *ast.CallExpr) string {
	if gf, ok := c.Callee.(*ast.GetFieldExpr); ok {
		if id, ok := gf.Object.(*ast.IdentExpr); ok && id.Name == BuiltinConsole {
			switch gf.Field {
			case "print":
				for _, arg := range c.Args {
					g.emit(Instr{Op: OpParam, Arg1: g.genExpr(arg)})
				}
				g.emit(Instr{Op: OpPrint, Arg2: fmt.Sprintf("%d", len(c.Args))})
				return ""
			case "readLine":
				for _, arg := range c.Args {
					g.emit(Instr{Op: OpParam, Arg1: g.genExpr(arg)})
				}
				t := g.freshTemp()
				g.emit(Instr{Op: OpReadLine, Result: t, Arg2: fmt.Sprintf("%d", len(c.Args))})
				return t
			case "readInt":
				t := g.freshTemp()
				g.emit(Instr{Op: OpReadInt, Result: t})
				return t
			}
		}
		if id, ok := gf.Object.(*ast.IdentExpr); ok && id.Name == BuiltinFile {
			switch gf.Field {
			case "read":
				path := g.genExpr(c.Args[0])
				t := g.freshTemp()
				g.emit(Instr{Op: OpFileRead, Result: t, Arg1: path})
				return t
			case "write":
				path := g.genExpr(c.Args[0])
				content := g.genExpr(c.Args[1])
				g.emit(Instr{Op: OpParam, Arg1: path})
				g.emit(Instr{Op: OpParam, Arg1: content})
				g.emit(Instr{Op: OpFileWrite, Arg2: "2"})
				return ""
			}
		}
		receiver := g.genExpr(gf.Object)
		g.emit(Instr{Op: OpParam, Arg1: receiver})
		for _, arg := range c.Args {
			g.emit(Instr{Op: OpParam, Arg1: g.genExpr(arg)})
		}
		className := g.resolveMethodClass(gf.Object, gf.Field)
		nArgs := 1 + len(c.Args)
		g.emit(Instr{Op: OpCall, Arg1: className + "." + gf.Field, Arg2: fmt.Sprintf("%d", nArgs)})
		t := g.freshTemp()
		g.emit(Instr{Op: OpAssign, Result: t, Arg1: "call_result", Comment: receiver + "." + gf.Field})
		return t
	}
	return g.freshTemp()
}

func (g *Generator) genField(f *ast.GetFieldExpr) string {
	obj := g.genExpr(f.Object)
	return obj + "." + f.Field
}

func (g *Generator) genLValue(e ast.Expr) string {
	switch ex := e.(type) {
	case *ast.IdentExpr:
		return ex.Name
	case *ast.GetFieldExpr:
		return g.genField(ex)
	case *ast.IndexExpr:
		obj := g.genExpr(ex.Object)
		idx := g.genExpr(ex.Index)
		return obj + "[" + idx + "]"
	default:
		return g.freshTemp()
	}
}

func (g *Generator) exprClass(e ast.Expr) string {
	if t := e.GetType(); t != nil {
		return t.Base
	}
	return "Object"
}

func (g *Generator) resolveMethodClass(e ast.Expr, method string) string {
	objClass := g.exprClass(e)
	if ci, ok := g.classes[objClass]; ok {
		for c := ci; c != nil; c = c.Super {
			if _, ok := c.Methods[method]; ok {
				return c.Name
			}
			if method == "glow" && c.Constructor != nil {
				return c.Name
			}
		}
	}
	return objClass
}

func (g *Generator) emit(i Instr) {
	g.instrs = append(g.instrs, i)
}

func (g *Generator) freshTemp() string {
	name := fmt.Sprintf("t%d", g.temp)
	g.temp++
	return name
}

func (g *Generator) freshLabel() string {
	name := fmt.Sprintf("L%d", g.label)
	g.label++
	return name
}

func tokenOp(t token.Type) string {
	return t.String()
}
