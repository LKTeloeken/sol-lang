package parser

import plexer "github.com/alecthomas/participle/v2/lexer"

// ── Top-level ─────────────────────────────────────────────────────────────────

type PProgram struct {
	Pos   plexer.Position
	Decls []*PTopDecl `@@*`
}

type PTopDecl struct {
	Pos    plexer.Position
	Class  *PClassDecl `(  @@`
	Import *PImport    `| @@`
	Alias  *PAlias     `| @@`
	Stmt   *PStmt      `| @@ )`
}

// ── Classes ───────────────────────────────────────────────────────────────────

type PClassDecl struct {
	Pos     plexer.Position
	Name    string     `Rise @Ident`
	Super   string     `( Radiate @Ident )?`
	Members []*PMember `"{" @@* "}"`
}

type PMember struct {
	Pos      plexer.Position
	GlowDecl *PGlowDecl  `(  @@`
	WithVis  *PVisMember `| @@ )`
}

// PVisMember covers both field declarations and ray (method) declarations,
// which both start with a visibility keyword.
type PVisMember struct {
	Pos   plexer.Position
	Vis   string      `@( Public | Private )`
	Ray   *PRayBody   `(  Ray @@`
	Field *PFieldBody `| @@ )`
}

type PGlowDecl struct {
	Pos    plexer.Position
	Params []*PParam `Glow "(" ( @@ ( "," @@ )* )? ")"`
	Body   *PBlock   `@@`
}

type PRayBody struct {
	Pos     plexer.Position
	Name    string    `@Ident`
	Params  []*PParam `"(" ( @@ ( "," @@ )* )? ")"`
	RetType *PType    `@@?`
	Body    *PBlock   `@@`
}

type PFieldBody struct {
	Pos  plexer.Position
	Type *PType `@@`
	Name string `@Ident ";"`
}

type PParam struct {
	Pos  plexer.Position
	Type *PType `@@`
	Name string `@Ident`
}

// ── Types ─────────────────────────────────────────────────────────────────────

type PType struct {
	Pos    plexer.Position
	Base   string   `@( IntType | FloatType | BoolType | StringType | VoidType | Ident )`
	Arrays []string `( "[" @"]" )*`
}

// ── Import / Alias ────────────────────────────────────────────────────────────

type PImport struct {
	Pos  plexer.Position
	Path string `Orbit @String ";"`
}

type PAlias struct {
	Pos  plexer.Position
	Name string `Star @Ident "="`
	Type *PType `@@ ";"`
}

// ── Statements ────────────────────────────────────────────────────────────────

type PBlock struct {
	Pos   plexer.Position
	Stmts []*PStmt `"{" @@* "}"`
}

type PStmt struct {
	Pos      plexer.Position
	Var      *PVarDecl      `(  @@`
	If       *PIfStmt       `| @@`
	While    *PWhileStmt    `| @@`
	For      *PForStmt      `| @@`
	Break    *PBreakStmt    `| @@`
	Continue *PContStmt     `| @@`
	Emit     *PEmitStmt     `| @@`
	Flare    *PFlareStmt    `| @@`
	Try      *PTryCatch     `| @@`
	Block    *PBlock        `| @@`
	Expr     *PAssignOrExpr `| @@ )`
}

type PVarDecl struct {
	Pos   plexer.Position
	Name  string `Var @Ident`
	Type  *PType `@@`
	Value *PExpr `( "=" @@ )? ";"`
}

type PIfStmt struct {
	Pos  plexer.Position
	Cond *PExpr  `If "(" @@ ")"`
	Then *PBlock `@@`
	Else *PBlock `( Else @@ )?`
}

type PWhileStmt struct {
	Pos  plexer.Position
	Cond *PExpr  `While "(" @@ ")"`
	Body *PBlock `@@`
}

type PForStmt struct {
	Pos   plexer.Position
	Each  *PForEach  `For ( @@`
	Range *PForRange `| @@ )`
}

type PForEach struct {
	Pos  plexer.Position
	Var  string  `Each @Ident In`
	Iter *PExpr  `@@`
	Body *PBlock `@@`
}

type PForRange struct {
	Pos   plexer.Position
	Var   string  `@Ident In`
	Start *PExpr  `@@`
	End   *PExpr  `DotDot @@`
	Body  *PBlock `@@`
}

type PBreakStmt struct {
	Pos plexer.Position `Break ";"`
}

type PContStmt struct {
	Pos plexer.Position `Continue ";"`
}

type PEmitStmt struct {
	Pos   plexer.Position
	Value *PExpr `Emit ( @@ )? ";"`
}

type PFlareStmt struct {
	Pos   plexer.Position
	Value *PExpr `Flare @@ ";"`
}

type PTryCatch struct {
	Pos      plexer.Position
	Try      *PBlock `Try @@`
	CatchVar string  `Catch "(" @Ident ")"`
	Catch    *PBlock `@@`
}

// PAssignOrExpr covers both assignment statements (x = expr;) and expression
// statements (expr;) — the Right pointer distinguishes them.
type PAssignOrExpr struct {
	Pos   plexer.Position
	Left  *PExpr `@@`
	Right *PExpr `( "=" @@ )? ";"`
}

// ── Expressions ───────────────────────────────────────────────────────────────

// Binary operator chains encoded as left-node + right-hand-side list so that
// participle (an LL parser) avoids left-recursion. The conversion step folds
// the list into left-associative ast.BinaryExpr nodes.

type PExpr struct {
	Pos  plexer.Position
	Left *PAndExpr `@@`
	Rest []*POrRHS `@@*`
}

type POrRHS struct {
	Right *PAndExpr `Or @@`
}

type PAndExpr struct {
	Pos  plexer.Position
	Left *PEqExpr  `@@`
	Rest []*PAndRHS `@@*`
}

type PAndRHS struct {
	Right *PEqExpr `And @@`
}

type PEqExpr struct {
	Pos  plexer.Position
	Left *PRelExpr `@@`
	Rest []*PEqRHS `@@*`
}

type PEqRHS struct {
	Op    string    `@( Eq | NotEq )`
	Right *PRelExpr `@@`
}

type PRelExpr struct {
	Pos  plexer.Position
	Left *PAddExpr  `@@`
	Rest []*PRelRHS `@@*`
}

type PRelRHS struct {
	Op    string    `@( LtEq | GtEq | "<" | ">" )`
	Right *PAddExpr `@@`
}

type PAddExpr struct {
	Pos  plexer.Position
	Left *PMulExpr  `@@`
	Rest []*PAddRHS `@@*`
}

type PAddRHS struct {
	Op    string    `@( "+" | "-" )`
	Right *PMulExpr `@@`
}

type PMulExpr struct {
	Pos  plexer.Position
	Left *PUnary   `@@`
	Rest []*PMulRHS `@@*`
}

type PMulRHS struct {
	Op    string  `@( "*" | "/" | "%" )`
	Right *PUnary `@@`
}

// PUnary is either a unary-op applied to another PUnary, or a postfix expression.
type PUnary struct {
	Pos  plexer.Position
	Op   *PUnaryOp `(  @@`
	Base *PPostfix `| @@ )`
}

type PUnaryOp struct {
	Pos     plexer.Position
	Sym     string  `@( "!" | "-" )`
	Operand *PUnary `@@`
}

// ── Postfix ───────────────────────────────────────────────────────────────────

type PPostfix struct {
	Pos  plexer.Position
	Base *PPrimary  `@@`
	Ops  []*PPostOp `@@*`
}

type PPostOp struct {
	Pos   plexer.Position
	Field *PFieldOp `(  @@`
	Index *PIndexOp `| @@`
	Call  *PCallOp  `| @@ )`
}

type PFieldOp struct {
	Pos  plexer.Position
	Name string   `"." @( Ident | Glow )`
	Call *PCallOp `@@?`
}

type PIndexOp struct {
	Pos   plexer.Position
	Index *PExpr `"[" @@ "]"`
}

type PCallOp struct {
	Pos  plexer.Position
	Args []*PExpr `"(" ( @@ ( "," @@ )* )? ")"`
}

// ── Primary ───────────────────────────────────────────────────────────────────

type PPrimary struct {
	Pos   plexer.Position
	Int   string      `(  @Int`
	Float string      `| @Float`
	Str   string      `| @String`
	Bool  string      `| @( True | False )`
	Null  string      `| @Null`
	This  string      `| @This`
	Super *PSuper     `| @@`
	New   *PNew       `| @@`
	Array *PArrayLit  `| @@`
	Paren *PParenExpr `| @@`
	Ident string      `| @Ident )`
}

type PSuper struct {
	Pos    plexer.Position
	Method string   `Radiate "." @( Ident | Glow )`
	Call   *PCallOp `@@?`
}

type PNew struct {
	Pos   plexer.Position
	Class string   `New @Ident`
	Args  []*PExpr `"(" ( @@ ( "," @@ )* )? ")"`
}

type PArrayLit struct {
	Pos   plexer.Position
	Items []*PExpr `"[" ( @@ ( "," @@ )* )? "]"`
}

type PParenExpr struct {
	Pos   plexer.Position
	Inner *PExpr `"(" @@ ")"`
}
