package ast

import "github.com/unisc/compiladores/sol/src/token"

// Pos holds source position.
type Pos struct {
	Line   int
	Column int
}

// TypeDesc represents a SOL type.
type TypeDesc struct {
	Pos      Pos
	Base     string // int, float, bool, string, void, or class name
	IsArray  bool
	ElemType *TypeDesc // for arrays
}

func (t *TypeDesc) String() string {
	if t == nil {
		return "void"
	}
	if t.IsArray {
		return t.ElemType.String() + "[]"
	}
	return t.Base
}

func (t *TypeDesc) Copy() *TypeDesc {
	if t == nil {
		return nil
	}
	c := *t
	if t.ElemType != nil {
		c.ElemType = t.ElemType.Copy()
	}
	return &c
}

// Node is implemented by all AST nodes.
type Node interface {
	Pos() Pos
}

// Program is the root node.
type Program struct {
	Decls []TopLevelDecl
}

func (p *Program) Pos() Pos {
	if len(p.Decls) > 0 {
		return p.Decls[0].Pos()
	}
	return Pos{}
}

// TopLevelDecl is a class or top-level statement.
type TopLevelDecl interface {
	Node
	topLevelDecl()
}

type ClassDecl struct {
	PosInfo   Pos
	Name      string
	SuperName string
	Members   []MemberDecl
}

func (c *ClassDecl) Pos() Pos { return c.PosInfo }
func (c *ClassDecl) topLevelDecl() {}

// ImportDecl imports another .sol file via orbit "path".
type ImportDecl struct {
	PosInfo Pos
	Path    string
}

func (i *ImportDecl) Pos() Pos       { return i.PosInfo }
func (i *ImportDecl) topLevelDecl() {}

// TypeAliasDecl defines a named type alias via star Name = Type;
type TypeAliasDecl struct {
	PosInfo Pos
	Name    string
	Type    *TypeDesc
}

func (t *TypeAliasDecl) Pos() Pos       { return t.PosInfo }
func (t *TypeAliasDecl) topLevelDecl() {}

type MemberDecl interface {
	Node
	memberDecl()
}

type FieldDecl struct {
	PosInfo    Pos
	Visibility string
	Type       *TypeDesc
	Name       string
}

func (f *FieldDecl) Pos() Pos { return f.PosInfo }
func (f *FieldDecl) memberDecl() {}

type Param struct {
	Type *TypeDesc
	Name string
	Pos  Pos
}

type GlowDecl struct {
	PosInfo Pos
	Params  []Param
	Body    *BlockStmt
}

func (g *GlowDecl) Pos() Pos { return g.PosInfo }
func (g *GlowDecl) memberDecl() {}

type RayDecl struct {
	PosInfo    Pos
	Visibility string
	Name       string
	Params     []Param
	ReturnType *TypeDesc
	Body       *BlockStmt
}

func (r *RayDecl) Pos() Pos { return r.PosInfo }
func (r *RayDecl) memberDecl() {}

// Statements

type Stmt interface {
	Node
	stmtNode()
}

type BlockStmt struct {
	PosInfo Pos
	Stmts   []Stmt
}

func (b *BlockStmt) Pos() Pos { return b.PosInfo }
func (b *BlockStmt) stmtNode() {}

type VarDeclStmt struct {
	PosInfo Pos
	Name    string
	Type    *TypeDesc
	Value   Expr
}

func (v *VarDeclStmt) Pos() Pos { return v.PosInfo }
func (v *VarDeclStmt) stmtNode() {}
func (v *VarDeclStmt) topLevelDecl() {}

type AssignStmt struct {
	PosInfo Pos
	Target  Expr
	Value   Expr
}

func (a *AssignStmt) Pos() Pos { return a.PosInfo }
func (a *AssignStmt) stmtNode() {}
func (a *AssignStmt) topLevelDecl() {}

type IfStmt struct {
	PosInfo Pos
	Cond    Expr
	Then    *BlockStmt
	Else    *BlockStmt
}

func (i *IfStmt) Pos() Pos { return i.PosInfo }
func (i *IfStmt) stmtNode() {}
func (i *IfStmt) topLevelDecl() {}

type WhileStmt struct {
	PosInfo Pos
	Cond    Expr
	Body    *BlockStmt
}

func (w *WhileStmt) Pos() Pos { return w.PosInfo }
func (w *WhileStmt) stmtNode() {}
func (w *WhileStmt) topLevelDecl() {}

type ForEachStmt struct {
	PosInfo Pos
	VarName string
	Iter    Expr
	Body    *BlockStmt
}

func (f *ForEachStmt) Pos() Pos { return f.PosInfo }
func (f *ForEachStmt) stmtNode() {}
func (f *ForEachStmt) topLevelDecl() {}

type ForRangeStmt struct {
	PosInfo Pos
	VarName string
	Start   Expr
	End     Expr
	Body    *BlockStmt
}

func (f *ForRangeStmt) Pos() Pos { return f.PosInfo }
func (f *ForRangeStmt) stmtNode() {}
func (f *ForRangeStmt) topLevelDecl() {}

type BreakStmt struct {
	PosInfo Pos
}

func (b *BreakStmt) Pos() Pos { return b.PosInfo }
func (b *BreakStmt) stmtNode() {}

type ContinueStmt struct {
	PosInfo Pos
}

func (c *ContinueStmt) Pos() Pos { return c.PosInfo }
func (c *ContinueStmt) stmtNode() {}

type EmitStmt struct {
	PosInfo Pos
	Value   Expr
}

func (e *EmitStmt) Pos() Pos { return e.PosInfo }
func (e *EmitStmt) stmtNode() {}

type FlareStmt struct {
	PosInfo Pos
	Value   Expr
}

func (f *FlareStmt) Pos() Pos { return f.PosInfo }
func (f *FlareStmt) stmtNode() {}
func (f *FlareStmt) topLevelDecl() {}

type TryCatchStmt struct {
	PosInfo  Pos
	Try      *BlockStmt
	CatchVar string
	Catch    *BlockStmt
}

func (t *TryCatchStmt) Pos() Pos { return t.PosInfo }
func (t *TryCatchStmt) stmtNode() {}
func (t *TryCatchStmt) topLevelDecl() {}

type ExprStmt struct {
	PosInfo Pos
	Expr    Expr
}

func (e *ExprStmt) Pos() Pos { return e.PosInfo }
func (e *ExprStmt) stmtNode() {}
func (e *ExprStmt) topLevelDecl() {}

// BaseExpr holds common expression fields.
type BaseExpr struct {
	PosInfo Pos
	Type    *TypeDesc
}

func (b BaseExpr) Pos() Pos              { return b.PosInfo }
func (b *BaseExpr) GetType() *TypeDesc   { return b.Type }
func (b *BaseExpr) SetType(t *TypeDesc)  { b.Type = t }

func NewBase(pos Pos) BaseExpr {
	return BaseExpr{PosInfo: pos}
}

// Expr is implemented by expression nodes.
type Expr interface {
	Node
	exprNode()
	GetType() *TypeDesc
	SetType(*TypeDesc)
}

type IntLit struct {
	BaseExpr
	Value int64
}

func (i *IntLit) exprNode() {}

type FloatLit struct {
	BaseExpr
	Value float64
}

func (f *FloatLit) exprNode() {}

type StringLit struct {
	BaseExpr
	Value string
}

func (s *StringLit) exprNode() {}

type BoolLit struct {
	BaseExpr
	Value bool
}

func (b *BoolLit) exprNode() {}

type NullLit struct {
	BaseExpr
}

func (n *NullLit) exprNode() {}

type IdentExpr struct {
	BaseExpr
	Name string
}

func (i *IdentExpr) exprNode() {}

type ThisExpr struct {
	BaseExpr
}

func (t *ThisExpr) exprNode() {}

type RadiateExpr struct {
	BaseExpr
}

func (e *RadiateExpr) exprNode() {}

type BinaryExpr struct {
	BaseExpr
	Left     Expr
	Operator token.Type
	Right    Expr
}

func (b *BinaryExpr) exprNode() {}

type UnaryExpr struct {
	BaseExpr
	Operator token.Type
	Operand  Expr
}

func (u *UnaryExpr) exprNode() {}

type CallExpr struct {
	BaseExpr
	Callee Expr
	Args   []Expr
}

func (c *CallExpr) exprNode() {}

type NewExpr struct {
	BaseExpr
	ClassName string
	Args      []Expr
}

func (n *NewExpr) exprNode() {}

type GetFieldExpr struct {
	BaseExpr
	Object Expr
	Field  string
}

func (g *GetFieldExpr) exprNode() {}

type ArrayLitExpr struct {
	BaseExpr
	Elements []Expr
}

func (a *ArrayLitExpr) exprNode() {}

type IndexExpr struct {
	BaseExpr
	Object Expr
	Index  Expr
}

func (i *IndexExpr) exprNode() {}

type ParenExpr struct {
	BaseExpr
	Inner Expr
}

func (p *ParenExpr) exprNode() {}

// SuperCallExpr represents radiate.glow(...) or radiate.method(...)
type SuperCallExpr struct {
	BaseExpr
	Method string
	Args   []Expr
}

func (s *SuperCallExpr) exprNode() {}
