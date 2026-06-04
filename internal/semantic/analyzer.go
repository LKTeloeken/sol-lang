package semantic

import (
	"fmt"

	"github.com/unisc/compiladores/sol/internal/ast"
	"github.com/unisc/compiladores/sol/internal/diag"
	"github.com/unisc/compiladores/sol/internal/stdlib"
	"github.com/unisc/compiladores/sol/internal/token"
)

type SymbolKind int

const (
	SymVar SymbolKind = iota
	SymParam
	SymField
	SymClass
)

type Symbol struct {
	Name string
	Kind SymbolKind
	Type *ast.TypeDesc
	Pos  ast.Pos
}

type Scope struct {
	symbols map[string]*Symbol
}

func NewScope() *Scope {
	return &Scope{symbols: make(map[string]*Symbol)}
}

func (s *Scope) Define(sym *Symbol) bool {
	if _, exists := s.symbols[sym.Name]; exists {
		return false
	}
	s.symbols[sym.Name] = sym
	return true
}

func (s *Scope) Lookup(name string) *Symbol {
	return s.symbols[name]
}

type MethodInfo struct {
	Name       string
	Visibility string
	Params     []ast.Param
	ReturnType *ast.TypeDesc
	Decl       *ast.RayDecl
}

type ClassInfo struct {
	Name       string
	SuperName  string
	Super      *ClassInfo
	Fields     map[string]*ast.FieldDecl
	Methods    map[string]*MethodInfo
	Constructor *ast.GlowDecl
	Decl       *ast.ClassDecl
}

type Analyzer struct {
	file    string
	classes map[string]*ClassInfo
	scopes  []*Scope
	errors  []diag.Error
	inRay   bool
	loopDepth int
	curClass *ClassInfo
}

func New(file string) *Analyzer {
	return &Analyzer{
		file:    file,
		classes: make(map[string]*ClassInfo),
	}
}

func (a *Analyzer) Errors() []diag.Error { return a.errors }

func (a *Analyzer) Check(prog *ast.Program) {
	a.registerBuiltins()
	a.collectClasses(prog)
	a.pushScope() // shared script scope for top-level variables
	for _, decl := range prog.Decls {
		switch d := decl.(type) {
		case *ast.ClassDecl:
			a.checkClass(d)
		default:
			a.checkTopLevelStmt(d)
		}
	}
	a.popScope()
}

func (a *Analyzer) registerBuiltins() {
	for _, bc := range stdlib.Classes {
		methods := make(map[string]*MethodInfo)
		for name, m := range bc.Methods {
			methods[name] = &MethodInfo{
				Name:       m.Name,
				Visibility: "public",
				ReturnType: m.ReturnType.Copy(),
			}
		}
		a.classes[bc.Name] = &ClassInfo{
			Name:    bc.Name,
			Methods: methods,
		}
	}
}

func (a *Analyzer) collectClasses(prog *ast.Program) {
	for _, decl := range prog.Decls {
		if c, ok := decl.(*ast.ClassDecl); ok {
			if stdlib.IsBuiltin(c.Name) {
				a.err(c.Pos(), "class name %q is reserved for stdlib", c.Name)
				continue
			}
			if _, exists := a.classes[c.Name]; exists {
				a.err(c.Pos(), "class %q already declared", c.Name)
				continue
			}
			info := &ClassInfo{
				Name:    c.Name,
				SuperName: c.SuperName,
				Fields:  make(map[string]*ast.FieldDecl),
				Methods: make(map[string]*MethodInfo),
				Decl:    c,
			}
			for _, m := range c.Members {
				switch mm := m.(type) {
				case *ast.FieldDecl:
					info.Fields[mm.Name] = mm
				case *ast.GlowDecl:
					info.Constructor = mm
				case *ast.RayDecl:
					info.Methods[mm.Name] = &MethodInfo{
						Name: mm.Name, Visibility: mm.Visibility,
						Params: mm.Params, ReturnType: mm.ReturnType, Decl: mm,
					}
				}
			}
			a.classes[c.Name] = info
		}
	}
	for _, info := range a.classes {
		if info.SuperName != "" {
			super, ok := a.classes[info.SuperName]
			if !ok {
				a.err(info.Decl.Pos(), "superclass %q not found for %q", info.SuperName, info.Name)
			} else {
				info.Super = super
			}
		}
	}
}

func (a *Analyzer) checkClass(c *ast.ClassDecl) {
	info := a.classes[c.Name]
	a.curClass = info
	a.pushScope()
	a.scopes[len(a.scopes)-1].Define(&Symbol{Name: "this", Kind: SymVar, Type: classType(c.Name), Pos: c.Pos()})
	if info.Super != nil {
		a.scopes[len(a.scopes)-1].Define(&Symbol{Name: "radiate", Kind: SymVar, Type: classType(info.SuperName), Pos: c.Pos()})
	}
	for _, m := range c.Members {
		switch mm := m.(type) {
		case *ast.GlowDecl:
			a.checkGlow(mm)
		case *ast.RayDecl:
			a.checkRay(mm)
		}
	}
	a.popScope()
	a.curClass = nil
}

func (a *Analyzer) checkGlow(g *ast.GlowDecl) {
	a.pushScope()
	for _, p := range g.Params {
		if !a.scopes[len(a.scopes)-1].Define(&Symbol{Name: p.Name, Kind: SymParam, Type: p.Type, Pos: p.Pos}) {
			a.err(p.Pos, "duplicate parameter %q", p.Name)
		}
	}
	a.inRay = true
	a.checkBlock(g.Body)
	a.inRay = false
	a.popScope()
}

func (a *Analyzer) checkRay(r *ast.RayDecl) {
	a.pushScope()
	for _, p := range r.Params {
		if !a.scopes[len(a.scopes)-1].Define(&Symbol{Name: p.Name, Kind: SymParam, Type: p.Type, Pos: p.Pos}) {
			a.err(p.Pos, "duplicate parameter %q", p.Name)
		}
	}
	a.inRay = true
	a.checkBlock(r.Body)
	if r.ReturnType != nil && r.ReturnType.Base != "void" {
		// simplified: skip return path analysis
	}
	a.inRay = false
	a.popScope()
}

func (a *Analyzer) checkTopLevelStmt(d ast.TopLevelDecl) {
	switch s := d.(type) {
	case *ast.VarDeclStmt:
		a.checkVarDecl(s, true)
	case *ast.AssignStmt:
		a.checkAssign(s)
	case *ast.IfStmt:
		a.checkIf(s)
	case *ast.WhileStmt:
		a.checkWhile(s)
	case *ast.ForEachStmt:
		a.checkForEach(s)
	case *ast.ForRangeStmt:
		a.checkForRange(s)
	case *ast.FlareStmt:
		a.err(s.Pos(), "flare at top level")
	case *ast.TryCatchStmt:
		a.checkTryCatch(s)
	case *ast.ExprStmt:
		a.checkExpr(s.Expr)
	}
}

func (a *Analyzer) checkBlock(b *ast.BlockStmt) {
	a.pushScope()
	for _, s := range b.Stmts {
		a.checkStmt(s)
	}
	a.popScope()
}

func (a *Analyzer) checkStmt(s ast.Stmt) {
	switch st := s.(type) {
	case *ast.VarDeclStmt:
		a.checkVarDecl(st, false)
	case *ast.AssignStmt:
		a.checkAssign(st)
	case *ast.IfStmt:
		a.checkIf(st)
	case *ast.WhileStmt:
		a.checkWhile(st)
	case *ast.ForEachStmt:
		a.checkForEach(st)
	case *ast.ForRangeStmt:
		a.checkForRange(st)
	case *ast.EmitStmt:
		if st.Value != nil {
			a.checkExpr(st.Value)
		}
	case *ast.FlareStmt:
		if !a.inRay {
			a.err(st.Pos(), "flare outside method")
		}
		a.checkExpr(st.Value)
	case *ast.TryCatchStmt:
		a.checkTryCatch(st)
	case *ast.ExprStmt:
		a.checkExpr(st.Expr)
	case *ast.BreakStmt:
		a.checkBreak(st)
	case *ast.ContinueStmt:
		a.checkContinue(st)
	case *ast.BlockStmt:
		a.checkBlock(st)
	}
}

func (a *Analyzer) checkVarDecl(v *ast.VarDeclStmt, topLevel bool) {
	if !a.scopes[len(a.scopes)-1].Define(&Symbol{Name: v.Name, Kind: SymVar, Type: v.Type, Pos: v.Pos()}) {
		a.err(v.Pos(), "variable %q already declared in scope", v.Name)
	}
	if v.Value != nil {
		vt := a.checkExpr(v.Value)
		if vt != nil && v.Type != nil && !typesCompatible(v.Type, vt) {
			a.err(v.Pos(), "cannot assign %s to %s", vt.String(), v.Type.String())
		}
	}
}

func (a *Analyzer) checkAssign(s *ast.AssignStmt) {
	lt := a.checkExpr(s.Target)
	rt := a.checkExpr(s.Value)
	if lt != nil && rt != nil && !typesCompatible(lt, rt) {
		a.err(s.Pos(), "cannot assign %s to %s", rt.String(), lt.String())
	}
}

func (a *Analyzer) checkIf(s *ast.IfStmt) {
	ct := a.checkExpr(s.Cond)
	if ct != nil && ct.Base != "bool" && ct.Base != "" {
		a.err(s.Pos(), "if condition must be bool")
	}
	a.checkBlock(s.Then)
	if s.Else != nil {
		a.checkBlock(s.Else)
	}
}

func (a *Analyzer) checkWhile(s *ast.WhileStmt) {
	ct := a.checkExpr(s.Cond)
	if ct != nil && ct.Base != "bool" && ct.Base != "" {
		a.err(s.Pos(), "while condition must be bool")
	}
	a.loopDepth++
	a.checkBlock(s.Body)
	a.loopDepth--
}

func (a *Analyzer) checkForEach(s *ast.ForEachStmt) {
	it := a.checkExpr(s.Iter)
	if it == nil || !it.IsArray {
		a.err(s.Pos(), "for each requires array expression")
	}
	a.pushScope()
	elemType := it.ElemType
	if elemType == nil {
		elemType = &ast.TypeDesc{Base: "void"}
	}
	a.scopes[len(a.scopes)-1].Define(&Symbol{Name: s.VarName, Kind: SymVar, Type: elemType, Pos: s.Pos()})
	a.loopDepth++
	a.checkBlock(s.Body)
	a.loopDepth--
	a.popScope()
}

func (a *Analyzer) checkForRange(s *ast.ForRangeStmt) {
	st := a.checkExpr(s.Start)
	et := a.checkExpr(s.End)
	if st != nil && st.Base != "int" {
		a.err(s.Start.Pos(), "for range start must be int")
	}
	if et != nil && et.Base != "int" {
		a.err(s.End.Pos(), "for range end must be int")
	}
	a.pushScope()
	a.scopes[len(a.scopes)-1].Define(&Symbol{Name: s.VarName, Kind: SymVar, Type: &ast.TypeDesc{Base: "int"}, Pos: s.Pos()})
	a.loopDepth++
	a.checkBlock(s.Body)
	a.loopDepth--
	a.popScope()
}

func (a *Analyzer) checkBreak(s *ast.BreakStmt) {
	if a.loopDepth == 0 {
		a.err(s.Pos(), "break outside loop")
	}
}

func (a *Analyzer) checkContinue(s *ast.ContinueStmt) {
	if a.loopDepth == 0 {
		a.err(s.Pos(), "continue outside loop")
	}
}

func (a *Analyzer) checkTryCatch(s *ast.TryCatchStmt) {
	a.checkBlock(s.Try)
	a.pushScope()
	a.scopes[len(a.scopes)-1].Define(&Symbol{Name: s.CatchVar, Kind: SymVar, Type: &ast.TypeDesc{Base: "string"}, Pos: s.Pos()})
	a.checkBlock(s.Catch)
	a.popScope()
}

func (a *Analyzer) checkExpr(e ast.Expr) *ast.TypeDesc {
	if e == nil {
		return nil
	}
	var result *ast.TypeDesc
	switch ex := e.(type) {
	case *ast.IntLit:
		result = &ast.TypeDesc{Base: "int"}
	case *ast.FloatLit:
		result = &ast.TypeDesc{Base: "float"}
	case *ast.StringLit:
		result = &ast.TypeDesc{Base: "string"}
	case *ast.BoolLit:
		result = &ast.TypeDesc{Base: "bool"}
	case *ast.NullLit:
		result = &ast.TypeDesc{Base: "null"}
	case *ast.IdentExpr:
		sym := a.lookup(ex.Name)
		if sym == nil {
			a.err(ex.Pos(), "undefined variable %q", ex.Name)
			result = &ast.TypeDesc{Base: "void"}
		} else {
			result = sym.Type
		}
	case *ast.ThisExpr:
		if a.curClass == nil {
			a.err(ex.Pos(), "this outside class")
			result = &ast.TypeDesc{Base: "void"}
		} else {
			result = classType(a.curClass.Name)
		}
	case *ast.RadiateExpr:
		if a.curClass == nil || a.curClass.Super == nil {
			a.err(ex.Pos(), "radiate used without superclass")
			result = &ast.TypeDesc{Base: "void"}
		} else {
			result = classType(a.curClass.SuperName)
		}
	case *ast.BinaryExpr:
		a.checkExpr(ex.Left)
		a.checkExpr(ex.Right)
		result = a.binaryType(ex)
	case *ast.UnaryExpr:
		a.checkExpr(ex.Operand)
		if ex.Operator == token.BANG {
			result = &ast.TypeDesc{Base: "bool"}
		} else {
			result = ex.Operand.GetType()
		}
	case *ast.CallExpr:
		result = a.checkCall(ex)
	case *ast.NewExpr:
		if _, ok := a.classes[ex.ClassName]; !ok {
			a.err(ex.Pos(), "unknown class %q", ex.ClassName)
		}
		for _, arg := range ex.Args {
			a.checkExpr(arg)
		}
		result = classType(ex.ClassName)
	case *ast.GetFieldExpr:
		result = a.checkField(ex)
	case *ast.ArrayLitExpr:
		var elem *ast.TypeDesc
		for _, el := range ex.Elements {
			t := a.checkExpr(el)
			if elem == nil {
				elem = t
			}
		}
		if elem == nil {
			elem = &ast.TypeDesc{Base: "void"}
		}
		result = &ast.TypeDesc{IsArray: true, ElemType: elem}
	case *ast.ParenExpr:
		result = a.checkExpr(ex.Inner)
	case *ast.IndexExpr:
		objType := a.checkExpr(ex.Object)
		idxType := a.checkExpr(ex.Index)
		if idxType != nil && idxType.Base != "int" && idxType.Base != "float" {
			a.err(ex.Pos(), "array index must be int")
		}
		if objType == nil || !objType.IsArray {
			a.err(ex.Pos(), "index on non-array type")
			result = &ast.TypeDesc{Base: "void"}
		} else if objType.ElemType != nil {
			result = objType.ElemType
		} else {
			result = &ast.TypeDesc{Base: "void"}
		}
	case *ast.SuperCallExpr:
		if a.curClass == nil || a.curClass.Super == nil {
			a.err(ex.Pos(), "super call without superclass")
		}
		for _, arg := range ex.Args {
			a.checkExpr(arg)
		}
		if mi := a.findMethod(a.curClass.Super, ex.Method); mi != nil {
			result = mi.ReturnType
		}
	}
	if result != nil {
		e.SetType(result)
	}
	return result
}

func (a *Analyzer) checkCall(c *ast.CallExpr) *ast.TypeDesc {
	if gf, ok := c.Callee.(*ast.GetFieldExpr); ok {
		if id, ok := gf.Object.(*ast.IdentExpr); ok && stdlib.IsBuiltin(id.Name) {
			return a.checkBuiltinCall(c, id.Name, gf.Field)
		}
		objType := a.checkExpr(gf.Object)
		for _, arg := range c.Args {
			a.checkExpr(arg)
		}
		if objType != nil && !objType.IsArray {
			ci := a.findClass(objType.Base)
			if ci == nil {
				a.err(c.Pos(), "call on non-class type")
				return &ast.TypeDesc{Base: "void"}
			}
			declClass := a.findDeclaringClass(ci, gf.Field)
			mi := a.findMethod(ci, gf.Field)
			if mi == nil {
				a.err(c.Pos(), "method %q not found on %s", gf.Field, objType.Base)
				return &ast.TypeDesc{Base: "void"}
			}
			if declClass != nil {
				a.checkMemberAccess(declClass, mi.Visibility, gf.Field, c.Pos(), true)
			}
			if len(c.Args) != len(mi.Params) {
				a.err(c.Pos(), "method %q expects %d args, got %d", gf.Field, len(mi.Params), len(c.Args))
			}
			if mi.ReturnType != nil {
				return mi.ReturnType
			}
			return &ast.TypeDesc{Base: "void"}
		}
	}
	a.checkExpr(c.Callee)
	for _, arg := range c.Args {
		a.checkExpr(arg)
	}
	return &ast.TypeDesc{Base: "void"}
}

func (a *Analyzer) checkBuiltinCall(c *ast.CallExpr, class, method string) *ast.TypeDesc {
	m, ok := stdlib.Lookup(class, method)
	if !ok {
		a.err(c.Pos(), "unknown static method %s.%s", class, method)
		return &ast.TypeDesc{Base: "void"}
	}
	n := len(c.Args)
	if m.MaxArgs < 0 {
		if n < m.MinArgs {
			a.err(c.Pos(), "%s.%s requires at least %d argument(s)", class, method, m.MinArgs)
		}
	} else {
		if n < m.MinArgs || n > m.MaxArgs {
			if m.MinArgs == m.MaxArgs {
				a.err(c.Pos(), "%s.%s requires %d argument(s), got %d", class, method, m.MinArgs, n)
			} else {
				a.err(c.Pos(), "%s.%s requires between %d and %d argument(s), got %d", class, method, m.MinArgs, m.MaxArgs, n)
			}
		}
	}
	for i, arg := range c.Args {
		t := a.checkExpr(arg)
		if m.PrintableArgs {
			if t != nil && !isPrintableType(t) {
				a.err(arg.Pos(), "Console.print argument must be int, float, bool, or string")
			}
			continue
		}
		if i < len(m.Args) {
			a.checkArgType(arg.Pos(), t, m.Args[i], class, method)
		}
	}
	if m.ReturnType != nil {
		return m.ReturnType.Copy()
	}
	return &ast.TypeDesc{Base: "void"}
}

func (a *Analyzer) checkArgType(pos ast.Pos, t *ast.TypeDesc, spec stdlib.ArgSpec, class, method string) {
	if t == nil {
		return
	}
	if spec.IsArray {
		if !t.IsArray || t.ElemType == nil || t.ElemType.Base != spec.ElemBase {
			a.err(pos, "%s.%s argument type mismatch: expected [%s]", class, method, spec.ElemBase)
		}
		return
	}
	if t.IsArray || t.Base != spec.Base {
		// allow int where float expected for Math
		if spec.Base == "float" && t.Base == "int" {
			return
		}
		a.err(pos, "%s.%s argument type mismatch: expected %s", class, method, spec.Base)
	}
}

func isPrintableType(t *ast.TypeDesc) bool {
	switch t.Base {
	case "int", "float", "bool", "string":
		return true
	default:
		return false
	}
}

func (a *Analyzer) checkField(g *ast.GetFieldExpr) *ast.TypeDesc {
	objType := a.checkExpr(g.Object)
	if objType == nil {
		return &ast.TypeDesc{Base: "void"}
	}
	if objType.IsArray {
		if g.Field == "length" {
			return &ast.TypeDesc{Base: "int"}
		}
		a.err(g.Pos(), "arrays only support .length field")
		return &ast.TypeDesc{Base: "void"}
	}
	if id, ok := g.Object.(*ast.IdentExpr); ok && stdlib.IsBuiltin(id.Name) {
		if ci, ok := a.classes[id.Name]; ok {
			if mi, ok := ci.Methods[g.Field]; ok {
				return mi.ReturnType
			}
		}
		a.err(g.Pos(), "unknown member %s.%s", id.Name, g.Field)
		return &ast.TypeDesc{Base: "void"}
	}
	ci := a.findClass(objType.Base)
	if ci == nil {
		a.err(g.Pos(), "field access on non-class type")
		return &ast.TypeDesc{Base: "void"}
	}
	if f, ok := ci.Fields[g.Field]; ok {
		declClass := a.findDeclaringFieldClass(ci, g.Field)
		if declClass != nil {
			a.checkMemberAccess(declClass, f.Visibility, g.Field, g.Pos(), false)
		}
		return f.Type
	}
	if mi := a.findMethod(ci, g.Field); mi != nil {
		return classType("function") // placeholder for method value
	}
	a.err(g.Pos(), "field %q not found on %s", g.Field, objType.Base)
	return &ast.TypeDesc{Base: "void"}
}

func (a *Analyzer) checkMemberAccess(declClass *ClassInfo, visibility, name string, pos ast.Pos, isMethod bool) {
	if visibility != "private" {
		return
	}
	if a.curClass == nil || a.curClass.Name != declClass.Name {
		kind := "field"
		if isMethod {
			kind = "method"
		}
		a.err(pos, "cannot access private %s %q from outside class %s", kind, name, declClass.Name)
	}
}

func (a *Analyzer) findDeclaringClass(ci *ClassInfo, name string) *ClassInfo {
	for c := ci; c != nil; c = c.Super {
		if _, ok := c.Methods[name]; ok {
			return c
		}
		if name == "glow" && c.Constructor != nil {
			return c
		}
	}
	return nil
}

func (a *Analyzer) findDeclaringFieldClass(ci *ClassInfo, name string) *ClassInfo {
	for c := ci; c != nil; c = c.Super {
		if _, ok := c.Fields[name]; ok {
			return c
		}
	}
	return nil
}

func (a *Analyzer) findClass(name string) *ClassInfo {
	if ci, ok := a.classes[name]; ok {
		return ci
	}
	return nil
}

func (a *Analyzer) findMethod(ci *ClassInfo, name string) *MethodInfo {
	for ci != nil {
		if mi, ok := ci.Methods[name]; ok {
			return mi
		}
		ci = ci.Super
	}
	return nil
}

func (a *Analyzer) binaryType(b *ast.BinaryExpr) *ast.TypeDesc {
	switch b.Operator {
	case token.PLUS:
		lt := b.Left.GetType()
		rt := b.Right.GetType()
		if isStringType(lt) || isStringType(rt) {
			if isStringCompatible(lt) && isStringCompatible(rt) {
				return &ast.TypeDesc{Base: "string"}
			}
			a.err(b.Pos(), "invalid operands for string concatenation")
			return &ast.TypeDesc{Base: "string"}
		}
		return &ast.TypeDesc{Base: "float"}
	case token.MINUS, token.ASTERISK, token.SLASH, token.PERCENT:
		return &ast.TypeDesc{Base: "float"}
	case token.EQ, token.NOT_EQ, token.LT, token.GT, token.LT_EQ, token.GT_EQ, token.AND, token.OR:
		return &ast.TypeDesc{Base: "bool"}
	default:
		return &ast.TypeDesc{Base: "void"}
	}
}

func isStringType(t *ast.TypeDesc) bool {
	return t != nil && t.Base == "string"
}

func isStringCompatible(t *ast.TypeDesc) bool {
	if t == nil {
		return true
	}
	switch t.Base {
	case "string", "int", "float", "bool":
		return true
	default:
		return false
	}
}

func (a *Analyzer) lookup(name string) *Symbol {
	for i := len(a.scopes) - 1; i >= 0; i-- {
		if sym := a.scopes[i].Lookup(name); sym != nil {
			return sym
		}
	}
	if a.curClass != nil {
		if f, ok := a.curClass.Fields[name]; ok {
			return &Symbol{Name: name, Kind: SymField, Type: f.Type, Pos: f.Pos()}
		}
	}
	return nil
}

func (a *Analyzer) pushScope() { a.scopes = append(a.scopes, NewScope()) }
func (a *Analyzer) popScope()  { a.scopes = a.scopes[:len(a.scopes)-1] }

func (a *Analyzer) err(pos ast.Pos, format string, args ...any) {
	a.errors = append(a.errors, diag.Error{
		File: a.file, Line: pos.Line, Column: pos.Column,
		Message: fmt.Sprintf(format, args...),
	})
}

func classType(name string) *ast.TypeDesc {
	return &ast.TypeDesc{Base: name}
}

func typesCompatible(a, b *ast.TypeDesc) bool {
	if a == nil || b == nil {
		return true
	}
	if a.Base == b.Base && a.IsArray == b.IsArray {
		if a.IsArray {
			return typesCompatible(a.ElemType, b.ElemType)
		}
		return true
	}
	if a.Base == "float" && b.Base == "int" {
		return true
	}
	if a.Base == "int" && b.Base == "float" {
		return true
	}
	if a.Base == "null" || b.Base == "null" {
		return true
	}
	return false
}

// Classes returns the class table for codegen.
func (a *Analyzer) Classes() map[string]*ClassInfo { return a.classes }
