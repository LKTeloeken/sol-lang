# SOL — Scripting Object Language

## Complete Language Specification & Interpreter Implementation Guide

> **Purpose of this document:** This file provides full context about the SOL programming language — its design rationale, syntax, grammar, and compiler implementation plan — intended to be used as context for AI assistants helping with development.

---

## 1. Origin & Concept

### The Wordplay

SOL is a scripting language designed as the **conceptual opposite of Lua**.

- **Lua** (Portuguese for _Moon_) is a lightweight, procedural/multi-paradigm scripting language widely used for embedded scripting in games and applications.
- **SOL** (Portuguese for _Sun_) is its counterpart: same scripting purpose, opposite paradigm.

The opposition is intentional and extends beyond the name:

| Aspect       | Lua                         | SOL                                   |
| ------------ | --------------------------- | ------------------------------------- |
| Name meaning | Moon                        | Sun                                   |
| Paradigm     | Procedural / multi-paradigm | Object-oriented                       |
| Type system  | Dynamic, duck-typed         | Static (type-checked at compile time) |
| Structure    | Functions and tables        | Classes, inheritance, encapsulation   |
| Syntax feel  | Minimal keywords, flexible  | Structured, explicit                  |

### Academic Context

SOL was designed as part of a **Formal Languages and Compilers** university course project. The goal is to create a complete compiler pipeline from scratch, implemented in **Go**, covering all classical compiler phases.

---

## 2. Language Paradigm

SOL is a **statically-typed, object-oriented scripting language**.

### Core OOP Features

- **Encapsulation** — class members are declared as `public` or `private`
- **Single inheritance** — a class can extend exactly one other class
- **Constructors** — each class has an optional constructor
- **Method dispatch** — methods are called on object instances
- **Exception handling** — structured try/catch with throw mechanism

### Design Philosophy

SOL is meant to feel familiar to developers who know Java or C#, but with a lighter, script-friendly syntax. There is no `main()` function — the program runs top-level statements directly, like a script. Classes are the primary organizational unit.

---

## 3. Sun-Themed Keywords

Every core structural keyword in SOL is a **solar metaphor**. This is both a design aesthetic and an easy way to identify SOL code at a glance.

| Keyword   | Role in Language             | Solar Metaphor                                |
| --------- | ---------------------------- | --------------------------------------------- |
| `rise`    | Declare a class              | A class rises — structure emerges at dawn     |
| `glow`    | Constructor of a class       | The initial glow when an object comes to life |
| `ray`     | Declare a method             | A ray of light emitted from the class         |
| `radiate` | Inheritance / call super   | A derived class radiates behavior from its parent |
| `emit`    | Return a value from a method | Emitting/radiating a result outward           |
| `flare`   | Throw an exception           | A solar flare — an unexpected burst of energy |

### General-Purpose Keywords (English, non-themed)

`public`, `private`, `var`, `new`, `this`, `if`, `else`, `while`, `for`, `each`, `in`, `break`, `continue`, `try`, `catch`, `true`, `false`, `null`

---

## 4. Syntax & Code Examples

### 4.1 Class Declaration (`rise`)

```sol
rise ContaBancaria {

    private float saldo;
    private string titular;

    glow(string titular, float saldoInicial) {
        this.titular = titular;
        this.saldo   = saldoInicial;
    }

    public ray depositar(float valor) {
        if (valor <= 0) {
            flare "Deposit value must be positive";
        }
        this.saldo = this.saldo + valor;
    }

    public ray sacar(float valor) {
        if (valor > this.saldo) {
            flare "Insufficient balance";
        }
        this.saldo = this.saldo - valor;
    }

    public ray getSaldo() {
        emit this.saldo;
    }

    public ray getTitular() {
        emit this.titular;
    }
}
```

### 4.2 Inheritance (`radiate`)

The keyword `radiate` is used both to **declare inheritance** and to **call the parent class** (equivalent to `super` in Java).

```sol
rise ContaEspecial radiate ContaBancaria {

    private float limiteCredito;

    glow(string titular, float saldoInicial, float limiteCredito) {
        radiate.glow(titular, saldoInicial);   // calls parent constructor
        this.limiteCredito = limiteCredito;
    }

    public ray sacar(float valor) {            // overrides parent method
        var disponivel = this.getSaldo() + this.limiteCredito;
        if (valor > disponivel) {
            flare "Total limit exceeded";
        }
        radiate.sacar(valor);                  // calls parent method
    }

    public ray getLimite() {
        emit this.limiteCredito;
    }
}
```

### 4.3 Object Instantiation & Usage

```sol
var conta ContaBancaria    = new ContaBancaria("Ana Silva", 1000.00);
var especial ContaEspecial = new ContaEspecial("Carlos Lima", 500.00, 300.00);

conta.depositar(250.00);
conta.sacar(100.00);
```

### 4.4 Control Flow

```sol
// if / else
if (conta.getSaldo() > 500) {
    conta.sacar(200.00);
} else {
    flare "Not enough balance for this operation";
}

// while loop
var i int = 0;
while (i < 10) {
    i = i + 1;
}

// for range — iterates over integer interval (semi-open)
for i in 0..10 {
    Console.print(i);  // prints 0..9
}

// for each — iterates over a list
var contas [ContaBancaria] = [conta, especial];
for each c in contas {
    var nome string  = c.getTitular();
    var saldo float  = c.getSaldo();
}
```

### 4.5 Exception Handling

```sol
try {
    conta.sacar(9999.00);
} catch (erro) {
    // handle the flare'd exception here
}
```

---

## 5. Formal Grammar (EBNF)

The grammar uses the following EBNF notation:

- `=` defines a production rule
- `;` ends a rule
- `|` means OR (alternative)
- `( )` groups elements
- `[ ]` means optional (0 or 1 time)
- `{ }` means repetition (0 or more times)
- `" "` denotes a literal terminal

### 5.1 Top-Level Structure

```ebnf
Program         = { TopLevelDecl } ;

TopLevelDecl    = ClassDecl | TopLevelStmt | ImportDecl ;

ImportDecl      = "orbit" StringLiteral ";" ;

TopLevelStmt    = VarDecl | AssignStmt | IfStmt | WhileStmt
                | ForEachStmt | ForRangeStmt | TryCatchStmt | ExprStmt ;

ClassDecl       = "rise" Identifier [ "radiate" Identifier ] "{" ClassBody "}" ;

ClassBody       = { MemberDecl } ;

MemberDecl      = FieldDecl
                | ConstructorDecl
                | MethodDecl ;
```

### 5.2 Class Members

```ebnf
FieldDecl       = ( "public" | "private" ) Type Identifier ";" ;

ConstructorDecl = "glow" "(" [ TypedParamList ] ")" Block ;

MethodDecl      = ( "public" | "private" ) "ray" Identifier "(" [ TypedParamList ] ")" [ Type ] Block ;

TypedParamList  = TypedParam { "," TypedParam } ;

TypedParam      = Type Identifier ;

ParamList       = Identifier { "," Identifier } ;
```

### 5.3 Statements

```ebnf
Block           = "{" { Statement } "}" ;

Statement       = VarDecl
                | AssignStmt
                | IfStmt
                | WhileStmt
                | ForEachStmt
                | ForRangeStmt
                | BreakStmt
                | ContinueStmt
                | ReturnStmt
                | FlareStmt
                | TryCatchStmt
                | ExprStmt ;

VarDecl         = "var" Identifier Type [ "=" Expression ] ";" ;

AssignStmt      = LValue "=" Expression ";" ;

LValue          = Identifier
                | "this" "." Identifier
                | "radiate" "." Identifier ;

IfStmt          = "if" "(" Expression ")" Block [ "else" Block ] ;

WhileStmt       = "while" "(" Expression ")" Block ;

ForEachStmt     = "for" "each" Identifier "in" Expression Block ;

ForRangeStmt    = "for" Identifier "in" Expression ".." Expression Block ;

BreakStmt       = "break" ";" ;

ContinueStmt    = "continue" ";" ;

ReturnStmt      = "emit" [ Expression ] ";" ;

FlareStmt       = "flare" Expression ";" ;

TryCatchStmt    = "try" Block "catch" "(" Identifier ")" Block ;

ExprStmt        = Expression ";" ;
```

### 5.4 Expressions (by precedence, lowest to highest)

```ebnf
Expression      = LogicalOr ;

LogicalOr       = LogicalAnd { "||" LogicalAnd } ;

LogicalAnd      = Equality { "&&" Equality } ;

Equality        = Relational { ( "==" | "!=" ) Relational } ;

Relational      = Additive { ( "<" | ">" | "<=" | ">=" ) Additive } ;

Additive        = Multiplicative { ( "+" | "-" ) Multiplicative } ;

Multiplicative  = Unary { ( "*" | "/" | "%" ) Unary } ;

Unary           = ( "!" | "-" ) Unary
                | Postfix ;

Postfix         = Primary { "." Identifier [ "(" [ ArgList ] ")" ] } ;

Primary         = IntLiteral
                | FloatLiteral
                | StringLiteral
                | BoolLiteral
                | NullLiteral
                | "this"
                | "radiate" "." "glow" "(" [ ArgList ] ")"
                | Identifier [ "(" [ ArgList ] ")" ]
                | "new" Identifier "(" [ ArgList ] ")"
                | "[" [ ExprList ] "]"
                | "(" Expression ")" ;

ArgList         = Expression { "," Expression } ;
ExprList        = Expression { "," Expression } ;
```

### 5.5 Types

```ebnf
Type            = BaseType | ClassType | AliasType | ArrayType ;

BaseType        = "int" | "float" | "bool" | "string" | "void" ;

ClassType       = Identifier ;

AliasType       = Identifier ;   /* resolved via star decl */

ArrayType       = Type "[]" ;

TypeAliasDecl   = "star" Identifier "=" Type ";" ;
```

| Type     | Syntax                       | Notes          |
| -------- | ---------------------------- | -------------- |
| `int`    | `private int saldo;`         | 64-bit signed  |
| `float`  | `private float limite;`      | 64-bit         |
| `bool`   | `private bool ativo;`        |                |
| `string` | `private string titular;`    |                |
| `void`   | method with no `emit`        |                |
| class    | `ContaBancaria`              | reference type |
| array    | `int[]`, `ContaBancaria[]`, `string[]` | postfix; `for each`, métodos |
| alias    | `star TodoItems = string[];`           | top-level only |

### 5.6 Terminals (Tokens)

```ebnf
Identifier      = Letter { Letter | Digit | "_" } ;

IntLiteral      = Digit { Digit } ;

FloatLiteral    = Digit { Digit } "." Digit { Digit } ;

StringLiteral   = '"' { Char } '"' ;

BoolLiteral     = "true" | "false" ;

NullLiteral     = "null" ;

LineComment     = "//" { any char except newline } ;

BlockComment    = "/*" { any char } "*/" ;

Letter          = "a" ... "z" | "A" ... "Z" | "_" ;
Digit           = "0" ... "9" ;
```

---

## 6. Token Types

The following is the complete list of token types used by the lexer, as they would be defined in Go:

```
// Literals
INT_LIT, FLOAT_LIT, STRING_LIT, BOOL_LIT, NULL_LIT

// Identifier
IDENT

// Sun-themed reserved words
rise, ray, glow, radiate, emit, flare

// General reserved words
public, private, var, new, this,
if, else, while, for, each, in, break, continue,
try, catch, true, false, null,
int, float, bool, string, void

// Operators
+  -  *  /  %  !  ..
== != <  >  <= >=
&& ||  =

// Delimiters
(  )  {  }  [  ]  .  ,  ;

// Control
EOF, ILLEGAL
```

---

## 7. Interpreter Implementation

SOL is implemented in **Go**, organized as a Go module with one package per phase. The project
originally targeted a native compiler (LLVM IR + `clang`); that backend was **removed**, leaving
the interpreted path: source → lexer → parser → module expansion → semantic → TAC → VM. The
parser/lexer use the [participle](https://github.com/alecthomas/participle) library.

### 7.1 Project Structure

```
sol-lang/
├── go.mod                    — module: github.com/unisc/compiladores/sol
├── Makefile                  — build, test, run, tac
├── src/
│   ├── main.go               — CLI driver (sollang)
│   ├── token/                — Token type + keywords (lexical definition)
│   ├── lexer/                — participle stateful lexer + token stream for -lex
│   ├── ast/                  — Node, Stmt, Expr interfaces
│   ├── parser/               — GLC via participle (grammar.go) + AST conversion (convert.go)
│   ├── semantic/             — symbol table, type checker, class/inheritance info
│   ├── modules/              — orbit import expansion
│   ├── tac/                  — Three-Address Code generator
│   ├── vm/                   — virtual machine that executes the TAC + builtins
│   ├── stdlib/               — static builtin registry (Console/File/Time/String/Math/Args)
│   ├── arraymethods/         — array method registry (push/pop/...)
│   ├── compiler/             — pipeline orchestration (driver)
│   └── diag/                 — diagnostics
├── examples/                 — valid SOL programs
└── testdata/invalid/         — programs that must fail
```

### 7.2 Execution Pipeline

```
source.sol
    │
    ▼  os.ReadFile()
participle lexer + parser → AST (Abstract Syntax Tree)
    │
    ▼
modules.Expand()         → inlines orbit imports
    │
    ▼
SemanticAnalyzer.Check() → annotates AST, validates types and scopes
    │
    ▼
tac.Generator.Build()    → emits Three-Address Code (TAC)
    │
    ▼
vm.VM.Run()              → interprets the TAC (default)
                           (or: -tac prints the TAC instead of running)
```

### 7.3 CLI Flags

The binary is named `sollang`. Running is the default; flags select earlier phases:

```bash
./sollang            source.sol    # compile to TAC and interpret (default)
./sollang -lex       source.sol    # run lexer only, print tokens
./sollang -parse     source.sol    # run lexer + parser, print AST summary
./sollang -check     source.sol    # run up to semantic analysis
./sollang -tac       source.sol    # emit TAC (stdout, or -o file)
./sollang -tac -o path source.sol  # write TAC to a file
```

---

## 8. Phase-by-Phase Implementation Guide

### Phase 01 — Formal Grammar Specification

**Goal:** Define the complete language before writing any code.

**Deliverables:**

- Complete EBNF grammar document (this file covers this)
- Table of all reserved words and their roles
- Operator precedence table
- At least 5 examples of valid SOL programs
- At least 5 examples of invalid SOL programs with explanation of why they are rejected

**Go output of this phase:** `lexer/token.go` — defines all `TokenType` constants and the `Keywords` map.

---

### Phase 02 — Lexical Analysis (Lexer)

**Goal:** Read raw source code character by character and produce a stream of tokens.

**Key concepts:**

- Tokens are defined declaratively as participle stateful-lexer rules in `lexer.Def`
  (`src/lexer/lexer.go`): floats/ints, strings, multi-char operators, then keywords (which must
  precede `Ident` so they win), identifiers and single-char punctuation.
- Whitespace and comments (`//` single-line, `/* */` multi-line) are elided by the parser.
- Keywords are matched with `\b` boundaries so `int` is a type but `integer` is an `Ident`.
- A thin wrapper (`Lexer.NextToken` / `Tokenize` + `mapToken`) re-maps participle tokens to the
  `token` package types — used to print the `-lex` token stream and to document the lexical spec.
- Unrecognized input becomes an `Illegal`/`ILLEGAL` token with line and column.

**Deliverables:** `src/token/token.go`, `src/lexer/lexer.go`, `src/lexer/lexer_test.go`

**Example — Token struct in Go:**

```go
type Token struct {
    Type    TokenType
    Lexeme  string
    Line    int
    Column  int
}
```

---

### Phase 03 — Syntactic Analysis (Parser)

**Goal:** Parse the source and build an Abstract Syntax Tree (AST).

**Key concepts:**

- Implementation strategy: **participle** — the GLC is declared as Go struct tags in
  `src/parser/grammar.go` (one struct per grammar production). participle generates the LL parser.
- A conversion pass (`src/parser/convert.go`) maps the participle parse tree into the internal AST
  in `src/ast/`, folding operator-chain lists into left-associative binary nodes.
- Operator precedence is encoded in the grammar layering (`PExpr → PAndExpr → … → PUnary → PPostfix → PPrimary`).
- The `radiate` keyword serves double duty: as inheritance marker in class declarations, and as super-call prefix inside method bodies — the grammar handles both contexts.
- Errors should be descriptive ("expected '{' after class name at line 5, column 12") and, when possible, the parser should attempt error recovery to report multiple errors in one pass.

**AST node examples:**

```go
// ShineDecl represents a class declaration
type ShineDecl struct {
    Name       string
    SuperClass string   // empty string if no radiate
    Fields     []FieldDecl
    Constructor *GlowDecl
    Methods    []RayDecl
}

// RayDecl represents a method declaration
type RayDecl struct {
    Name       string
    Visibility string   // "public" or "private"
    Params     []string
    Body       *Block
}
```

**Deliverables:** `src/ast/ast.go`, `src/parser/grammar.go`, `src/parser/convert.go`, `src/parser/parser.go`, `src/parser/parser_test.go`

---

### Phase 04 — Semantic Analysis

**Goal:** Walk the AST and verify that the program is semantically correct — not just syntactically valid.

**Key concepts:**

- **Symbol Table:** a stack of scopes. Each scope is a `map[string]Symbol`. Entering a block pushes a new scope; leaving pops it.
- **Visitor Pattern:** the semantic analyzer implements the `Visitor` interface and walks every node in the AST.

**Checks to implement:**

1. No variable used before declaration
2. No variable declared twice in the same scope
3. All method calls reference methods that exist on the target class
4. `radiate` in a class body refers to an actually declared class
5. `emit` expression type matches the expected return type of the enclosing `ray`
6. `flare` is used inside a method body (not at top level)
7. `new ClassName()` — `ClassName` must be a declared `rise`
8. Argument count matches parameter count in method calls

**Deliverables:** `semantic/symbol_table.go`, `semantic/type_checker.go`, `semantic/semantic_test.go`

---

### Phase 05 — Intermediate Code Generation

**Goal:** Translate the semantically-validated AST into Three-Address Code (TAC), a low-level intermediate representation.

**Key concepts:**

- TAC is a sequence of simple instructions, each with at most 3 operands: `result = operand1 op operand2`
- Temporary variables (`t0`, `t1`, `t2`, …) are generated as needed
- Method calls become `call ClassName.methodName, argCount` instructions
- `flare` becomes a `throw` instruction; `try/catch` becomes labeled jump targets

**Example TAC output** for `conta.depositar(250.00)`:

```
t0 = 250.00
param t0
call ContaBancaria.depositar, 1
```

**Example TAC output** for `if (valor <= 0) { flare "..."; }`:

```
t0 = valor <= 0
ifFalse t0 goto L1
throw "Deposit value must be positive"
L1:
```

**Deliverables:** `src/tac/generator.go`, `src/tac/generator_test.go`

---

### Phase 07 — LLVM Backend (Removed)

> This phase was an experimental native backend (TAC → LLVM IR → `clang` → binary). It has been
> **removed** from the project, which now ships only the interpreted path (TAC executed by the VM).

---

### Phase 06 — Interpreter Driver & CLI

**Goal:** Wire all phases together into a working command-line interpreter.

**Key concepts:**

- `src/main.go` reads the source file, calls `compiler.RunFileWithOptions`, and propagates errors.
- Any phase can return a list of errors — all errors are collected and printed together before stopping, so the user sees all problems at once, not just the first one.
- Error output format: `[ERROR] filename.sol:line:column — message`
- The `-lex`, `-parse`, `-check`, `-tac` flags stop the pipeline at a specific phase (useful for
  debugging and grading); with no flag the program is executed by the VM.

**Makefile targets:**

```makefile
build:
	go build -o sollang ./src/main.go

test:
	go test ./...

run: build
	./sollang examples/conta_bancaria.sol

tac: build
	./sollang -tac examples/conta_bancaria.sol

clean:
	rm -f sollang output.tac
```

**Deliverables:** `src/main.go`, `src/compiler/driver.go`, `Makefile`, `examples/*.sol`, test suite

---

## 9. Summary Table

| Phase | Name                  | Input            | Output                     | Go Packages                |
| ----- | --------------------- | ---------------- | -------------------------- | -------------------------- |
| 01    | Grammar Specification | —                | EBNF doc + token constants | `src/token/`               |
| 02    | Lexical Analysis      | Source string    | `[]Token`                  | `src/lexer/`               |
| 03    | Syntactic Analysis    | Source string    | AST                        | `src/ast/`, `src/parser/`  |
| 04    | Semantic Analysis     | AST              | Annotated AST              | `src/semantic/`            |
| 05    | Code Generation       | Annotated AST    | TAC instructions           | `src/tac/`                 |
| 06    | Execution             | TAC instructions | Program output             | `src/vm/`                  |
| 07    | Driver & CLI          | Source file path | Output / TAC               | `src/main.go`, `src/compiler/` |

---

## 10. Key Design Decisions

**Why Go?**
Go compiles to a single binary and has an excellent standard library for string processing and I/O. Its explicit interfaces are a natural fit for the type switches used across multiple phases.

**Why participle (lexer + parser)?**
The assignment suggests using a ready-made lexer/parser library. participle builds a stateful
lexer and an LL parser directly from struct-tag grammar rules, so the GLC lives next to the AST
types and stays easy to verify by inspection. Left-recursion is avoided by encoding binary
operator chains as left-node + RHS lists, folded into left-associative nodes during AST
conversion (`src/parser/convert.go`).

**Why TAC as the final output?**
Three-Address Code is the standard intermediate representation taught in compilers courses. It is human-readable, easy to verify manually, and close enough to assembly to demonstrate code generation concepts without requiring a full machine code backend.

**Why typed operands and explicit field/index instructions?**
Each TAC operand is a typed value — a constant or a name (temporary/variable) — rather than a raw
string, and composite accesses (`obj.f`, `a[i]`, array length) are their own instructions
(`fieldGet`/`fieldSet`/`indexGet`/`indexSet`/`len`). This keeps the IR unambiguous and lets the VM
execute each instruction by a direct lookup/assignment instead of re-parsing operand text at
runtime. An earlier version encoded these as strings (`"obj.f"`, `"a[i]"`) that the VM split back
apart — convenient to emit, but fragile and not proper three-address code; the explicit form is the
textbook representation and makes both the generator and the VM simpler.

**Why single inheritance only?**
Multiple inheritance introduces diamond-problem complexity that is out of scope for a compiler course project. The `radiate` keyword enforces exactly one parent class, keeping the class hierarchy simple and the semantic analysis tractable.

**Why static typing?**
A statically-typed language allows type errors to be caught at compile time (in the semantic phase) rather than at runtime, which is a better pedagogical choice for demonstrating semantic analysis. It also makes the type checker significantly simpler to implement correctly.

**Why a TAC VM instead of a native backend?**
An experimental LLVM/`clang` backend was prototyped and then removed: it added a C runtime, CGO-free
IR generation and an external toolchain dependency for little pedagogical gain. Interpreting the TAC
with a small VM keeps the project to a single Go binary while still demonstrating code generation.

**Why script-style top-level statements?**
SOL has no explicit `main()` — programs run top-level statements directly. The grammar allows
classes and statements at program root; the TAC generator wraps top-level code in a `__program`
block that the VM runs after loading the classes.

---

_End of SOL Language Specification — version 1.1_
