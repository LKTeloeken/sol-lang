# SOL — Scripting Object Language

## Complete Language Specification & Compiler Implementation Guide

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
Type            = BaseType | ClassType | ArrayType ;

BaseType        = "int" | "float" | "bool" | "string" | "void" ;

ClassType       = Identifier ;

ArrayType       = "[" Type "]" ;
```

| Type     | Syntax                       | Notes          |
| -------- | ---------------------------- | -------------- |
| `int`    | `private int saldo;`         | 64-bit signed  |
| `float`  | `private float limite;`      | 64-bit         |
| `bool`   | `private bool ativo;`        |                |
| `string` | `private string titular;`    |                |
| `void`   | method with no `emit`        |                |
| class    | `ContaBancaria`              | reference type |
| array    | `[ContaBancaria]` or `[int]` | for `for each` |

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

## 7. Compiler Implementation Plan

The SOL compiler is implemented in **Go**, organized as a standard Go module with one package per compiler phase.

### 7.1 Project Structure

```
sol/
├── go.mod                    — module: github.com/unisc/compiladores/sol
├── Makefile                  — build, test, run commands
├── cmd/solc/main.go          — CLI driver
├── internal/
│   ├── token/                — TokenType enum, Token struct, Keywords map
│   ├── lexer/                — Lexer struct, NextToken() method
│   ├── ast/                  — Node, Statement, Expression interfaces
│   ├── parser/               — recursive descent + Pratt parser
│   ├── semantic/             — symbol table, type checker
│   ├── tac/                  — Three-Address Code generator
│   └── llvm/                 — LLVM IR generator (bonus backend)
├── runtime/solrt.c           — minimal runtime (panic, helpers)
├── examples/                 — valid SOL programs
└── testdata/                 — valid + invalid test programs
```

### 7.2 Execution Pipeline

```
source.sol
    │
    ▼  io.ReadFile()
Lexer.NextToken()        → produces []Token
    │
    ▼
Parser.Parse()           → produces AST (Abstract Syntax Tree)
    │
    ▼
SemanticAnalyzer.Check() → annotates AST, validates types and scopes
    │
    ▼
Generator.Generate()     → emits Three-Address Code (TAC)
    │
    ▼
output.tac
    │
    ▼  (optional)
LLVMGen.Generate()       → emits LLVM IR
    │
    ▼
clang output.ll runtime/solrt.c -o program
```

### 7.3 CLI Flags

The compiled binary is named `solc` and supports the following flags:

```bash
./solc --lex      source.sol    # run lexer only, print tokens
./solc --parse    source.sol    # run lexer + parser, print AST
./solc --check    source.sol    # run up to semantic analysis
./solc --compile  source.sol    # full pipeline, emit TAC output
./solc --emit-ir  source.sol    # emit LLVM IR to output.ll
./solc --build    source.sol    # TAC → LLVM IR → native binary
./solc -o path    source.sol    # custom output path
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

- The `Lexer` struct holds the source string, current position, current character, line, and column.
- `NextToken()` is called repeatedly to advance through the source.
- Whitespace and comments (`//` single-line, `/* */` multi-line) are skipped silently.
- When an identifier is scanned, look it up in the `Keywords` map — if found, emit the keyword token; otherwise emit `IDENT`.
- Errors produce an `ILLEGAL` token with a descriptive message, line, and column.

**Deliverables:** `lexer/lexer.go`, `lexer/lexer_test.go`

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

**Goal:** Consume the token stream from the Lexer and build an Abstract Syntax Tree (AST).

**Key concepts:**

- Implementation strategy: **Recursive Descent Parser** — one function per grammar rule.
- Each grammar production maps directly to a Go function: `parseProgram()`, `parseClassDecl()`, `parseMethodDecl()`, `parseStatement()`, `parseExpression()`, etc.
- Expression parsing uses **Pratt parsing** (top-down operator precedence) to handle the precedence hierarchy correctly.
- The `radiate` keyword serves double duty: as inheritance marker in class declarations, and as super-call prefix inside method bodies — the parser must handle both contexts.
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

**Deliverables:** `ast/nodes.go`, `ast/declarations.go`, `ast/visitor.go`, `parser/parser.go`, `parser/parser_test.go`

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

**Deliverables:** `internal/tac/instruction.go`, `internal/tac/generator.go`, `internal/tac/generator_test.go`

---

### Phase 07 — LLVM Backend (Bonus)

**Goal:** Lower the annotated AST to LLVM IR and produce a native executable.

**Key concepts:**

- Use `llir/llvm` (pure Go) to generate LLVM IR without CGO
- Invoke system `clang` to link IR with `runtime/solrt.c`
- Top-level statements are wrapped in an implicit `@main` function
- `int` → `i64`, `float` → `double`, `bool` → `i1`, `string` → `{ i8*, i64 }`
- `flare` → call `@sol_panic(i8* msg)`; v1 uses static dispatch

**Deliverables:** `internal/llvm/irgen.go`, `runtime/solrt.c`

---

### Phase 06 — Compiler Driver & CLI

**Goal:** Wire all phases together into a working command-line compiler.

**Key concepts:**

- `main.go` reads the source file, calls each phase in sequence, and propagates errors.
- Any phase can return a list of errors — all errors are collected and printed together before stopping, so the user sees all problems at once, not just the first one.
- Error output format: `[ERROR] filename.sol:line:column — message`
- The `--lex`, `--parse`, `--check` flags allow stopping the pipeline at a specific phase (useful for debugging and grading).

**Makefile targets:**

```makefile
build:
	go build -o solc .

test:
	go test ./...

run:
	./solc --compile examples/conta_bancaria.sol

clean:
	rm -f solc
```

**Deliverables:** `main.go`, `cmd/solc.go`, `Makefile`, `examples/*.sol`, integration test suite

---

## 9. Summary Table

| Phase | Name                  | Input            | Output                     | Go Packages       |
| ----- | --------------------- | ---------------- | -------------------------- | ----------------- |
| 01    | Grammar Specification | —                | EBNF doc + token constants | `lexer/token.go`  |
| 02    | Lexical Analysis      | Source string    | `[]Token`                  | `lexer/`          |
| 03    | Syntactic Analysis    | `[]Token`        | AST                        | `ast/`, `parser/` |
| 04    | Semantic Analysis     | AST              | Annotated AST              | `semantic/`       |
| 05    | Code Generation       | Annotated AST    | TAC instructions           | `internal/tac/`   |
| 06    | Compiler Driver       | Source file path | TAC output file            | `cmd/solc/`       |
| 07    | LLVM Backend          | Annotated AST    | Native binary              | `internal/llvm/`  |

---

## 10. Key Design Decisions

**Why Go?**
Go has no dependency on external parser frameworks, compiles to a single binary, and has an excellent standard library for string processing and I/O. Its explicit interfaces are a natural fit for the Visitor pattern used across multiple compiler phases.

**Why Recursive Descent?**
The grammar is LL(1)-friendly (with minimal lookahead needed). Recursive descent is readable, debuggable, and maps one-to-one with the EBNF grammar rules — making it easy to verify correctness by inspection.

**Why TAC as the final output?**
Three-Address Code is the standard intermediate representation taught in compilers courses. It is human-readable, easy to verify manually, and close enough to assembly to demonstrate code generation concepts without requiring a full machine code backend.

**Why single inheritance only?**
Multiple inheritance introduces diamond-problem complexity that is out of scope for a compiler course project. The `radiate` keyword enforces exactly one parent class, keeping the class hierarchy simple and the semantic analysis tractable.

**Why static typing?**
A statically-typed language allows type errors to be caught at compile time (in the semantic phase) rather than at runtime, which is a better pedagogical choice for demonstrating semantic analysis. It also makes the type checker significantly simpler to implement correctly.

**Why LLVM?**
LLVM provides a production-grade backend: optimization passes, multi-platform code generation, and a well-documented IR. The frontend (lexer through semantic analysis) stays in pure Go; LLVM handles machine code so the project can produce runnable binaries beyond TAC.

**Why llir/llvm?**
Official LLVM Go bindings were removed from upstream LLVM (2022). `llir/llvm` is a pure Go library for generating LLVM IR — no CGO, no compiling LLVM from source. External tools (`clang`, `llc`) compile the emitted `.ll` files.

**Why script-style top-level statements?**
SOL has no explicit `main()` — programs run top-level statements directly. The grammar allows classes and statements at program root; LLVM codegen wraps top-level code in an implicit `@main`.

---

_End of SOL Language Specification — version 1.1_
