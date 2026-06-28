# SOL (`sollang`)

Interpretador da linguagem **SOL** (*Scripting Object Language*) — uma DSL de script,
estaticamente tipada e orientada a objetos.

SOL é o trabalho da disciplina de **Linguagens Formais e Compiladores** (UNISC). O projeto
cobre as etapas clássicas de um interpretador/gerador de código:

- **Análise léxica** — tokens definidos no pacote [`src/token`](src/token/) e reconhecidos por um
  lexer *stateful* do [participle](https://github.com/alecthomas/participle).
- **Análise sintática** — gramática livre de contexto (GLC) declarada em
  [`src/parser/grammar.go`](src/parser/grammar.go), também via participle.
- **Análise semântica** — verificação de tipos, classes e herança em
  [`src/semantic`](src/semantic/).
- **Saída útil** — o programa é **interpretado** por uma máquina virtual e, opcionalmente, pode
  ter o **código intermediário (TAC)** gerado e impresso (`-tac`).

Para aprender a **escrever programas SOL** (sintaxe, tipos, classes, operadores, exemplos),
consulte **[`SYNTAX.md`](SYNTAX.md)**.

---

## Estrutura do projeto

```
sol-lang/
├── src/
│   ├── main.go            # Executável CLI (sollang)
│   ├── token/             # Definição de tokens
│   ├── lexer/             # Análise léxica (participle stateful lexer)
│   ├── parser/            # Análise sintática: GLC (grammar.go) + AST (convert.go)
│   ├── ast/               # Árvore sintática abstrata
│   ├── semantic/          # Análise semântica (tipos, classes, herança)
│   ├── modules/           # Expansão de imports (orbit)
│   ├── tac/               # Geração de código intermediário (três endereços)
│   ├── vm/                # Máquina virtual que executa o TAC
│   ├── stdlib/            # Builtins estáticos (Console, File, Time, String, Math, Args)
│   ├── arraymethods/      # Métodos de array (push, pop, ...)
│   ├── compiler/          # Orquestração do pipeline (driver)
│   └── diag/              # Mensagens de erro
├── examples/              # Programas de exemplo (.sol)
├── testdata/invalid/      # Casos que devem falhar (erros de léxico/sintaxe/semântica)
├── Makefile               # Atalhos de build e execução
├── go.mod                 # Módulo Go local
└── SYNTAX.md              # Referência completa da linguagem
```

O módulo Go é **local** — o caminho `github.com/unisc/compiladores/sol` em [`go.mod`](go.mod) não
precisa existir no GitHub; todos os pacotes estão neste repositório.

---

## Pré-requisitos

| Ferramenta | Versão | Obrigatório para |
|------------|--------|------------------|
| [Go](https://go.dev/dl/) | 1.22 ou superior | Compilar e rodar o `sollang` |

```bash
go version
# go version go1.22.x ...
```

---

## Passo a passo: compilar e rodar

Todos os comandos assumem que você está na **raiz do repositório** (onde ficam `go.mod` e
`Makefile`).

### 1. Compilar o interpretador

```bash
make build
# ou diretamente:
go build -o sollang ./src/main.go
```

O executável `./sollang` será criado na raiz do projeto.

### 2. Executar um programa SOL

Rodar (interpretar) é o comportamento **padrão** — basta passar o arquivo:

```bash
./sollang examples/hello.sol
```

Saída esperada:

```
Hello, SOL
```

Outros exemplos prontos:

```bash
./sollang examples/conta_bancaria.sol
./sollang examples/heranca.sol
./sollang examples/for_each.sol
./sollang examples/control_flow.sol
./sollang examples/modules/main.sol      # usa orbit (imports)
```

### 3. Rodar seu próprio programa

Crie um arquivo `.sol` e execute:

```bash
./sollang meu_programa.sol
```

Use `Console.print(...)` para ver saída no terminal. Sintaxe completa em [`SYNTAX.md`](SYNTAX.md).

### 4. Atalho com Make

```bash
make run    # build + ./sollang examples/conta_bancaria.sol
```

---

## Comandos do `sollang`

| Comando | O que faz |
|---------|-----------|
| `./sollang arquivo.sol` | Compila e **executa** o script (padrão) |
| `./sollang -lex arquivo.sol` | Apenas análise léxica (lista de tokens) |
| `./sollang -parse arquivo.sol` | Apenas parsing (resumo do AST) |
| `./sollang -check arquivo.sol` | Apenas análise semântica (tipos) |
| `./sollang -tac arquivo.sol` | Gera e imprime o código intermediário (TAC) |

Salvar o TAC em arquivo (a flag `-o` vem **antes** do arquivo `.sol`):

```bash
./sollang -tac -o output.tac programa.sol
```

Argumentos para o script vêm **depois** do arquivo `.sol` (acessíveis via `Args`):

```bash
./sollang meu_programa.sol arg1 arg2
```

---

## Testes

```bash
make test
# ou
go test ./...
```

Limpar artefatos gerados (`sollang`, `output.tac`):

```bash
make clean
```

---

## Ferramentas e dependências

Conforme sugerido no enunciado, a análise léxica e sintática usa uma **biblioteca pronta** em vez
de ser escrita à mão:

- [`participle/v2`](https://github.com/alecthomas/participle) — lexer *stateful* e parser a partir
  de uma GLC declarativa (tags de struct em [`src/parser/grammar.go`](src/parser/grammar.go)).
- A stdlib da linguagem delega para a biblioteca padrão do Go: `math` (`Math.*`),
  `strings` (`String.*`), `time` (`Time.*`) e `os` (`File.*`).

---

## Aprender a programar em SOL

Consulte **[`SYNTAX.md`](SYNTAX.md)** — referência completa com palavras reservadas
(`rise`, `glow`, `ray`, `radiate`, `emit`, `flare`, …), tipos, operadores, controle de fluxo,
classes/herança/exceções, arrays, `for each`/`for i in 0..10`, stdlib e exemplos.

Exemplo mínimo:

```sol
rise Foo {
    private int x;

    glow(int x) {
        this.x = x;
    }

    public ray getX() int {
        emit this.x;
    }
}

var f Foo = new Foo(42);
Console.print(f.getX());
```
