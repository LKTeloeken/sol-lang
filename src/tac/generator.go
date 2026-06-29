package tac

import (
	"fmt"
	"strings"

	"github.com/unisc/compiladores/sol/src/arraymethods"
	"github.com/unisc/compiladores/sol/src/ast"
	"github.com/unisc/compiladores/sol/src/semantic"
	"github.com/unisc/compiladores/sol/src/stdlib"
	"github.com/unisc/compiladores/sol/src/token"
)

// ── Operands ────────────────────────────────────────────────────────────────
//
// A TAC operand is one of three things, distinguished by its kind:
//   - a constant literal (int/float/bool/string/null),
//   - a name (temporary or variable), or
//   - the pending result of the most recent call.
//
// Encoding operands as a typed value (instead of a raw string) means the VM
// never has to re-parse text at runtime to decide whether something is a
// literal, a variable, a field access or an array index.

type OperandKind int

const (
	OpndNone       OperandKind = iota // empty operand (e.g. a bare `return`)
	OpndConst                         // a literal value
	OpndName                          // a temporary or variable name
	OpndCallResult                    // the return value of the last call
)

// LitKind is the type of a constant operand.
type LitKind int

const (
	LitNull LitKind = iota
	LitInt
	LitFloat
	LitBool
	LitStr
)

// Operand is a typed three-address-code operand.
type Operand struct {
	Kind  OperandKind
	Lit   LitKind // valid when Kind == OpndConst
	Int   int64
	Float float64
	Bool  bool
	Str   string
	Name  string // valid when Kind == OpndName
}

func ConstInt(v int64) Operand     { return Operand{Kind: OpndConst, Lit: LitInt, Int: v} }
func ConstFloat(v float64) Operand { return Operand{Kind: OpndConst, Lit: LitFloat, Float: v} }
func ConstBool(v bool) Operand     { return Operand{Kind: OpndConst, Lit: LitBool, Bool: v} }
func ConstStr(v string) Operand    { return Operand{Kind: OpndConst, Lit: LitStr, Str: v} }
func ConstNull() Operand           { return Operand{Kind: OpndConst, Lit: LitNull} }
func Name(n string) Operand        { return Operand{Kind: OpndName, Name: n} }
func CallResult() Operand          { return Operand{Kind: OpndCallResult} }

// IsZero reports whether the operand is unset (used to detect a bare `return`).
func (o Operand) IsZero() bool { return o.Kind == OpndNone }

func (o Operand) String() string {
	switch o.Kind {
	case OpndConst:
		switch o.Lit {
		case LitNull:
			return "null"
		case LitInt:
			return fmt.Sprintf("%d", o.Int)
		case LitFloat:
			return fmt.Sprintf("%g", o.Float)
		case LitBool:
			if o.Bool {
				return "true"
			}
			return "false"
		case LitStr:
			return fmt.Sprintf("%q", o.Str)
		}
	case OpndName:
		return o.Name
	case OpndCallResult:
		return "call_result"
	}
	return "_"
}

// ── Instructions ────────────────────────────────────────────────────────────

type Op string

const (
	OpAssign    Op = "assign"
	OpBinOp     Op = "binop"
	OpUnary     Op = "unary"
	OpLabel     Op = "label"
	OpGoto      Op = "goto"
	OpIfFalse   Op = "ifFalse"
	OpParam     Op = "param"
	OpCall      Op = "call"
	OpThrow     Op = "throw"
	OpNew       Op = "new"
	OpReturn    Op = "return"
	OpComment   Op = "comment"
	OpBeginTry  Op = "beginTry"
	OpEndTry    Op = "endTry"
	OpBuiltin   Op = "builtin"
	OpArrayLit  Op = "arrayLit"
	OpArrayCall Op = "arrayCall"
	OpFieldGet  Op = "fieldGet"
	OpFieldSet  Op = "fieldSet"
	OpIndexGet  Op = "indexGet"
	OpIndexSet  Op = "indexSet"
	OpLen       Op = "len"
)

type loopContext struct {
	breakLabel    string
	continueLabel string
}

// Instr is a single three-address-code instruction. Only the fields relevant to
// each Op are populated (see Generator for which fields each op uses).
type Instr struct {
	Op       Op
	Dst      string  // destination temp/variable name (value-producing ops)
	A        Operand // first input operand
	B        Operand // second input operand (binops)
	Obj      Operand // object/array receiver (field/index/array-method ops)
	Idx      Operand // index operand (index ops)
	Field    string  // field name (field get/set)
	Operator string  // operator symbol (binop/unary)
	Sym      string  // label / class / method / builtin id / catch var
	Label    string  // jump target label
	NArgs    int     // argument count (call/builtin/arrayLit/arrayCall)
	Comment  string
}

func (i Instr) String() string {
	switch i.Op {
	case OpAssign:
		if i.Comment != "" {
			return fmt.Sprintf("%s = %s  ; %s", i.Dst, i.A, i.Comment)
		}
		return fmt.Sprintf("%s = %s", i.Dst, i.A)
	case OpBinOp:
		return fmt.Sprintf("%s = %s %s %s", i.Dst, i.A, i.Operator, i.B)
	case OpUnary:
		return fmt.Sprintf("%s = %s%s", i.Dst, i.Operator, i.A)
	case OpFieldGet:
		return fmt.Sprintf("%s = %s.%s", i.Dst, i.Obj, i.Field)
	case OpFieldSet:
		return fmt.Sprintf("%s.%s = %s", i.Obj, i.Field, i.A)
	case OpIndexGet:
		return fmt.Sprintf("%s = %s[%s]", i.Dst, i.Obj, i.Idx)
	case OpIndexSet:
		return fmt.Sprintf("%s[%s] = %s", i.Obj, i.Idx, i.A)
	case OpLen:
		return fmt.Sprintf("%s = len(%s)", i.Dst, i.Obj)
	case OpLabel:
		return fmt.Sprintf("%s:", i.Label)
	case OpGoto:
		return fmt.Sprintf("goto %s", i.Label)
	case OpIfFalse:
		return fmt.Sprintf("ifFalse %s goto %s", i.A, i.Label)
	case OpParam:
		return fmt.Sprintf("param %s", i.A)
	case OpCall:
		return fmt.Sprintf("call %s, %d", i.Sym, i.NArgs)
	case OpThrow:
		return fmt.Sprintf("throw %s", i.A)
	case OpNew:
		return fmt.Sprintf("%s = new %s", i.Dst, i.Sym)
	case OpReturn:
		if !i.A.IsZero() {
			return fmt.Sprintf("return %s", i.A)
		}
		return "return"
	case OpComment:
		return "; " + i.Comment
	case OpBeginTry:
		return fmt.Sprintf("beginTry %s %s", i.Label, i.Sym)
	case OpEndTry:
		return fmt.Sprintf("endTry %s", i.Label)
	case OpBuiltin:
		if i.Dst != "" {
			return fmt.Sprintf("%s = builtin %s %d", i.Dst, i.Sym, i.NArgs)
		}
		return fmt.Sprintf("builtin %s %d", i.Sym, i.NArgs)
	case OpArrayLit:
		return fmt.Sprintf("%s = arrayLit %d", i.Dst, i.NArgs)
	case OpArrayCall:
		if i.Dst != "" {
			return fmt.Sprintf("%s = arrayCall %s %s %d", i.Dst, i.Obj, i.Sym, i.NArgs)
		}
		return fmt.Sprintf("arrayCall %s %s %d", i.Obj, i.Sym, i.NArgs)
	default:
		return fmt.Sprintf("; unknown op %s", i.Op)
	}
}

// ── Generator ───────────────────────────────────────────────────────────────

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
	return g.Format()
}

// Format renders the already-built instruction list as TAC text.
func (g *Generator) Format() string {
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
			g.emit(Instr{Op: OpReturn, A: g.genExpr(st.Value)})
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
		g.emit(Instr{Op: OpAssign, Dst: v.Name, A: g.genExpr(v.Value)})
	}
}

func (g *Generator) genAssign(s *ast.AssignStmt) {
	val := g.genExpr(s.Value)
	switch t := s.Target.(type) {
	case *ast.IdentExpr:
		g.emit(Instr{Op: OpAssign, Dst: t.Name, A: val})
	case *ast.GetFieldExpr:
		obj := g.genExpr(t.Object)
		g.emit(Instr{Op: OpFieldSet, Obj: obj, Field: t.Field, A: val})
	case *ast.IndexExpr:
		obj := g.genExpr(t.Object)
		idx := g.genExpr(t.Index)
		g.emit(Instr{Op: OpIndexSet, Obj: obj, Idx: idx, A: val})
	}
}

func (g *Generator) genIf(s *ast.IfStmt) {
	cond := g.genExpr(s.Cond)
	elseLabel := g.freshLabel()
	endLabel := g.freshLabel()
	if s.Else == nil {
		endLabel = elseLabel
	}
	g.emit(Instr{Op: OpIfFalse, A: cond, Label: elseLabel})
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
	g.emit(Instr{Op: OpIfFalse, A: cond, Label: end})
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
	g.emit(Instr{Op: OpAssign, Dst: idx, A: ConstInt(0), Comment: "for each index"})
	loopStart := g.freshLabel()
	loopEnd := g.freshLabel()
	contLabel := g.freshLabel()
	g.pushLoop(loopEnd, contLabel)
	g.emit(Instr{Op: OpLabel, Label: loopStart})
	g.emit(Instr{Op: OpLen, Dst: lenTmp, Obj: arr, Comment: "array length"})
	lt := g.freshTemp()
	g.emit(Instr{Op: OpBinOp, Dst: lt, A: Name(idx), Operator: "<", B: Name(lenTmp)})
	g.emit(Instr{Op: OpIfFalse, A: Name(lt), Label: loopEnd})
	g.emit(Instr{Op: OpIndexGet, Dst: s.VarName, Obj: arr, Idx: Name(idx), Comment: "for each elem"})
	for _, st := range s.Body.Stmts {
		g.genStmt(st)
	}
	g.emit(Instr{Op: OpLabel, Label: contLabel})
	inc := g.freshTemp()
	g.emit(Instr{Op: OpBinOp, Dst: inc, A: Name(idx), Operator: "+", B: ConstInt(1)})
	g.emit(Instr{Op: OpAssign, Dst: idx, A: Name(inc)})
	g.emit(Instr{Op: OpGoto, Label: loopStart})
	g.emit(Instr{Op: OpLabel, Label: loopEnd, Comment: "end for each"})
	g.popLoop()
}

func (g *Generator) genForRange(s *ast.ForRangeStmt) {
	startVal := g.genExpr(s.Start)
	endVal := g.genExpr(s.End)
	g.emit(Instr{Op: OpAssign, Dst: s.VarName, A: startVal, Comment: "for range init"})
	loopStart := g.freshLabel()
	loopEnd := g.freshLabel()
	contLabel := g.freshLabel()
	g.pushLoop(loopEnd, contLabel)
	g.emit(Instr{Op: OpLabel, Label: loopStart})
	lt := g.freshTemp()
	g.emit(Instr{Op: OpBinOp, Dst: lt, A: Name(s.VarName), Operator: "<", B: endVal})
	g.emit(Instr{Op: OpIfFalse, A: Name(lt), Label: loopEnd})
	for _, st := range s.Body.Stmts {
		g.genStmt(st)
	}
	g.emit(Instr{Op: OpLabel, Label: contLabel})
	inc := g.freshTemp()
	g.emit(Instr{Op: OpBinOp, Dst: inc, A: Name(s.VarName), Operator: "+", B: ConstInt(1)})
	g.emit(Instr{Op: OpAssign, Dst: s.VarName, A: Name(inc)})
	g.emit(Instr{Op: OpGoto, Label: loopStart})
	g.emit(Instr{Op: OpLabel, Label: loopEnd, Comment: "end for range"})
	g.popLoop()
}

func (g *Generator) genFlare(s *ast.FlareStmt) {
	g.emit(Instr{Op: OpThrow, A: g.genExpr(s.Value)})
}

func (g *Generator) genTryCatch(s *ast.TryCatchStmt) {
	catchLabel := g.freshLabel()
	endLabel := g.freshLabel()
	g.emit(Instr{Op: OpBeginTry, Label: catchLabel, Sym: s.CatchVar})
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

// genExpr lowers an expression, emitting any necessary instructions, and returns
// the operand that holds its value.
func (g *Generator) genExpr(e ast.Expr) Operand {
	switch ex := e.(type) {
	case *ast.IntLit:
		return ConstInt(ex.Value)
	case *ast.FloatLit:
		return ConstFloat(ex.Value)
	case *ast.StringLit:
		return ConstStr(ex.Value)
	case *ast.BoolLit:
		return ConstBool(ex.Value)
	case *ast.NullLit:
		return ConstNull()
	case *ast.IdentExpr:
		return Name(ex.Name)
	case *ast.ThisExpr:
		return Name("this")
	case *ast.RadiateExpr:
		return Name("radiate")
	case *ast.ParenExpr:
		return g.genExpr(ex.Inner)
	case *ast.BinaryExpr:
		a := g.genExpr(ex.Left)
		b := g.genExpr(ex.Right)
		t := g.freshTemp()
		g.emit(Instr{Op: OpBinOp, Dst: t, A: a, Operator: tokenOp(ex.Operator), B: b})
		return Name(t)
	case *ast.UnaryExpr:
		a := g.genExpr(ex.Operand)
		op := "-"
		if ex.Operator == token.BANG {
			op = "!"
		}
		t := g.freshTemp()
		g.emit(Instr{Op: OpUnary, Dst: t, Operator: op, A: a})
		return Name(t)
	case *ast.CallExpr:
		return g.genCall(ex)
	case *ast.NewExpr:
		t := g.freshTemp()
		g.emit(Instr{Op: OpNew, Dst: t, Sym: ex.ClassName})
		g.emit(Instr{Op: OpParam, A: Name(t)})
		for _, arg := range ex.Args {
			g.emit(Instr{Op: OpParam, A: g.genExpr(arg)})
		}
		g.emit(Instr{Op: OpCall, Sym: ex.ClassName + ".glow", NArgs: 1 + len(ex.Args)})
		return Name(t)
	case *ast.GetFieldExpr:
		obj := g.genExpr(ex.Object)
		t := g.freshTemp()
		if ot := ex.Object.GetType(); ot != nil && ot.IsArray && ex.Field == "length" {
			g.emit(Instr{Op: OpLen, Dst: t, Obj: obj})
		} else {
			g.emit(Instr{Op: OpFieldGet, Dst: t, Obj: obj, Field: ex.Field})
		}
		return Name(t)
	case *ast.SuperCallExpr:
		g.emit(Instr{Op: OpParam, A: Name("this")})
		for _, arg := range ex.Args {
			g.emit(Instr{Op: OpParam, A: g.genExpr(arg)})
		}
		superClass := ""
		if g.curClass != "" {
			if ci, ok := g.classes[g.curClass]; ok && ci.Super != nil {
				superClass = ci.SuperName
			}
		}
		g.emit(Instr{Op: OpCall, Sym: superClass + "." + ex.Method, NArgs: 1 + len(ex.Args)})
		t := g.freshTemp()
		g.emit(Instr{Op: OpAssign, Dst: t, A: CallResult(), Comment: "super " + ex.Method})
		return Name(t)
	case *ast.ArrayLitExpr:
		for _, el := range ex.Elements {
			g.emit(Instr{Op: OpParam, A: g.genExpr(el)})
		}
		t := g.freshTemp()
		g.emit(Instr{Op: OpArrayLit, Dst: t, NArgs: len(ex.Elements)})
		return Name(t)
	case *ast.IndexExpr:
		obj := g.genExpr(ex.Object)
		idx := g.genExpr(ex.Index)
		t := g.freshTemp()
		g.emit(Instr{Op: OpIndexGet, Dst: t, Obj: obj, Idx: idx})
		return Name(t)
	default:
		return Name(g.freshTemp())
	}
}

func (g *Generator) genCall(c *ast.CallExpr) Operand {
	gf, ok := c.Callee.(*ast.GetFieldExpr)
	if !ok {
		return Name(g.freshTemp())
	}

	// Static stdlib builtin, e.g. Console.print(...)
	if id, ok := gf.Object.(*ast.IdentExpr); ok && stdlib.IsBuiltin(id.Name) {
		if _, ok := stdlib.Lookup(id.Name, gf.Field); ok {
			for _, arg := range c.Args {
				g.emit(Instr{Op: OpParam, A: g.genExpr(arg)})
			}
			ins := Instr{Op: OpBuiltin, Sym: stdlib.BuiltinID(id.Name, gf.Field), NArgs: len(c.Args)}
			if stdlib.ReturnsVoid(id.Name, gf.Field) {
				g.emit(ins)
				return Operand{}
			}
			t := g.freshTemp()
			ins.Dst = t
			g.emit(ins)
			return Name(t)
		}
	}

	// Array method, e.g. arr.push(x)
	if m, ok := arraymethods.Lookup(gf.Field); ok {
		if ot := gf.Object.GetType(); ot != nil && ot.IsArray {
			recv, writeback := g.genArrayReceiver(gf.Object)
			for _, arg := range c.Args {
				g.emit(Instr{Op: OpParam, A: g.genExpr(arg)})
			}
			ins := Instr{Op: OpArrayCall, Obj: Name(recv), Sym: gf.Field, NArgs: len(c.Args)}
			var ret Operand
			if m.ReturnType == "void" {
				g.emit(ins)
			} else {
				d := g.freshTemp()
				ins.Dst = d
				g.emit(ins)
				ret = Name(d)
			}
			if writeback != nil {
				writeback()
			}
			return ret
		}
	}

	// User-defined method, e.g. obj.method(args)
	receiver := g.genExpr(gf.Object)
	g.emit(Instr{Op: OpParam, A: receiver})
	for _, arg := range c.Args {
		g.emit(Instr{Op: OpParam, A: g.genExpr(arg)})
	}
	className := g.resolveMethodClass(gf.Object, gf.Field)
	g.emit(Instr{Op: OpCall, Sym: className + "." + gf.Field, NArgs: 1 + len(c.Args)})
	t := g.freshTemp()
	g.emit(Instr{Op: OpAssign, Dst: t, A: CallResult(), Comment: gf.Field})
	return Name(t)
}

// genArrayReceiver loads an array l-value into an addressable name. For a plain
// variable the name is returned directly; for a field or index access the array
// is loaded into a temporary and a write-back closure stores it back after a
// mutating array method runs.
func (g *Generator) genArrayReceiver(e ast.Expr) (string, func()) {
	switch ex := e.(type) {
	case *ast.IdentExpr:
		return ex.Name, nil
	case *ast.GetFieldExpr:
		obj := g.genExpr(ex.Object)
		t := g.freshTemp()
		g.emit(Instr{Op: OpFieldGet, Dst: t, Obj: obj, Field: ex.Field})
		return t, func() { g.emit(Instr{Op: OpFieldSet, Obj: obj, Field: ex.Field, A: Name(t)}) }
	case *ast.IndexExpr:
		obj := g.genExpr(ex.Object)
		idx := g.genExpr(ex.Index)
		t := g.freshTemp()
		g.emit(Instr{Op: OpIndexGet, Dst: t, Obj: obj, Idx: idx})
		return t, func() { g.emit(Instr{Op: OpIndexSet, Obj: obj, Idx: idx, A: Name(t)}) }
	default:
		v := g.genExpr(e)
		if v.Kind == OpndName {
			return v.Name, nil
		}
		t := g.freshTemp()
		g.emit(Instr{Op: OpAssign, Dst: t, A: v})
		return t, nil
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
