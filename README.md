# Compilador SOL (`solc`)

Compilador da linguagem **SOL** (*Scripting Object Language*) — linguagem de script estaticamente tipada e orientada a objetos.

Para aprender a **escrever programas SOL** (sintaxe, tipos, classes, operadores, exemplos), consulte **[`SYNTAX.md`](SYNTAX.md)**.

---

## Estrutura do projeto

```
sol/
├── cmd/solc/              # Executável CLI do compilador
├── internal/
│   ├── token/             # Definição de tokens
│   ├── lexer/             # Análise léxica
│   ├── parser/            # Análise sintática (AST)
│   ├── ast/               # Árvore sintática abstrata
│   ├── semantic/          # Análise semântica (tipos)
│   ├── tac/               # Geração de código intermediário (TAC)
│   ├── vm/                # Máquina virtual que executa o TAC
│   ├── compiler/          # Orquestração do pipeline + LLVM IR
│   └── diag/              # Mensagens de erro
├── runtime/               # Runtime C (usado com --build)
├── examples/              # Programas de exemplo (.sol)
├── testdata/              # Casos de teste inválidos
├── vscode-extension/      # Extensão de syntax highlighting (VS Code / Cursor)
├── editor/nvim/           # Syntax highlighting para Neovim / Vim
├── Makefile               # Atalhos de build e execução
├── go.mod                 # Módulo Go local
└── SYNTAX.md              # Referência completa da linguagem
```

O módulo Go é **local** — o caminho `github.com/unisc/compiladores/sol` em [`go.mod`](go.mod) não precisa existir no GitHub; todos os pacotes estão neste repositório.

---

## Pré-requisitos

| Ferramenta | Versão | Obrigatório para |
|------------|--------|------------------|
| [Go](https://go.dev/dl/) | 1.22 ou superior | Compilar o `solc` |
| [clang](https://releases.llvm.org/) | qualquer recente | Apenas `--build` (binário nativo) |

Verifique se o Go está instalado:

```bash
go version
# go version go1.22.x ...
```

---

## Passo a passo: compilar e rodar

### 1. Clonar ou entrar no diretório do projeto

```bash
cd /caminho/para/sol
```

Todos os comandos abaixo assumem que você está na **raiz do repositório** (onde ficam `go.mod` e `Makefile`).

### 2. Compilar o compilador

**Opção A — com Make (recomendado):**

```bash
make build
```

**Opção B — com Go diretamente:**

```bash
go build -o solc ./cmd/solc
```

Se tudo correr bem, o executável `./solc` será criado na raiz do projeto.

### 3. Executar um programa SOL

A forma mais comum é compilar e **interpretar** o script com a VM interna:

```bash
./solc --run examples/hello.sol
```

Saída esperada:

```
Hello, SOL
program finished OK
```

Outros exemplos prontos:

```bash
./solc --run examples/simple.sol
./solc --run examples/script.sol
./solc --run examples/conta_bancaria.sol
./solc --run examples/heranca.sol
./solc --run examples/for_each.sol
./solc --run examples/control_flow.sol
```

### 4. Rodar seu próprio programa

Crie um arquivo `.sol` (por exemplo `meu_programa.sol`) e execute:

```bash
./solc --run meu_programa.sol
```

Use `Console.print(...)` para ver saída no terminal. Veja a sintaxe completa em [`SYNTAX.md`](SYNTAX.md).

### 5. Atalho com Make

O `Makefile` já compila e executa um exemplo:

```bash
make run
```

Equivalente a `make build` seguido de `./solc --run examples/conta_bancaria.sol`.

---

## Comandos do `solc`

| Comando | O que faz |
|---------|-----------|
| `./solc --run arquivo.sol` | Compila e **executa** o script (VM TAC) |
| `./solc --compile arquivo.sol` | Gera código intermediário TAC (padrão: `output.tac`) |
| `./solc --check arquivo.sol` | Apenas análise semântica (tipos) |
| `./solc --parse arquivo.sol` | Apenas parsing (AST) |
| `./solc --lex arquivo.sol` | Apenas lexer (lista de tokens) |
| `./solc --emit-ir arquivo.sol` | Gera LLVM IR (padrão: `output.ll`) |
| `./solc --build arquivo.sol` | Gera IR + compila binário nativo com `clang` |

Definir arquivo de saída:

```bash
./solc --compile programa.sol -o saida.tac
./solc --emit-ir programa.sol -o saida.ll
./solc --build programa.sol -o programa.ll   # gera programa (binário)
```

Executar o binário nativo (requer `clang`):

```bash
./solc --build examples/hello.sol -o hello.ll
./hello
```

> O backend nativo (`--build`) é **experimental**. Para desenvolvimento e testes, prefira `./solc --run`.

---

## Testes

Rodar a suíte de testes do compilador:

```bash
make test
# ou
go test ./...
```

Limpar artefatos gerados:

```bash
make clean
```

Remove `solc`, `output.tac`, `output.ll` e `program`.

---

## Aprender a programar em SOL

Consulte **[`SYNTAX.md`](SYNTAX.md)** — referência completa com:

- Palavras reservadas (`rise`, `glow`, `ray`, `enlights`, `emit`, `flare`, …)
- Tipos, operadores e controle de fluxo
- Classes, herança e exceções
- Arrays, `for each` e `for i in 0..10`
- Stdlib: `Console.print`, `Console.readLine`, `Console.readInt`, `File.read`, `File.write`
- `break` e `continue`
- Exemplos completos e limitações atuais

Exemplo mínimo (detalhes em `SYNTAX.md`):

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

---

## Editor (syntax highlighting)

Arquivos SOL usam a extensão `.sol`.

| Editor | Pasta | Instalação |
|--------|-------|------------|
| VS Code / Cursor | [`vscode-extension/`](vscode-extension/) | `cd vscode-extension && npm install && npm run package` → instalar o `.vsix` |
| Neovim / Vim | [`editor/nvim/`](editor/nvim/) | Symlink de `syntax/` e `ftdetect/` para sua config |

Se o VS Code abrir `.sol` como Solidity, adicione em `settings.json`:

```json
"files.associations": { "*.sol": "sol" }
```

---

## Roadmap

Limitações conhecidas e plano de evolução: [`PLANO-LIMITACOES.md`](PLANO-LIMITACOES.md).
