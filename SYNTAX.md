# SOL — Referência de Sintaxe

Guia prático da linguagem **SOL** (*Scripting Object Language*): tipos, keywords, operadores, OOP e como usar o compilador.

> SOL é estaticamente tipada, orientada a objetos e executada como **script** (sem função `main`). O compilador `solc` analisa o código e a VM interpreta o TAC gerado.

---

## Sumário

1. [Visão geral](#visão-geral)
2. [Palavras reservadas](#palavras-reservadas)
3. [Tipos](#tipos)
4. [Literais e comentários](#literais-e-comentários)
5. [Classes e OOP](#classes-e-oop)
6. [Variáveis e atribuição](#variáveis-e-atribuição)
7. [Operadores](#operadores)
8. [Controle de fluxo](#controle-de-fluxo)
9. [Exceções (`flare` / `try` / `catch`)](#exceções-flare--try--catch)
10. [Programa script (top-level)](#programa-script-top-level)
11. [Como executar](#como-executar)
12. [Logs e depuração](#logs-e-depuração)
13. [Limitações atuais](#limitações-atuais)
14. [Exemplos completos](#exemplos-completos)

---

## Visão geral

| Característica | SOL |
|----------------|-----|
| Paradigma | Orientada a objetos |
| Tipagem | Estática (erros no compile-time) |
| Herança | Simples (uma classe pai) |
| Entrada do programa | Statements no nível raiz |
| Extensão de arquivo | `.sol` |
| Compilador | `solc` |

SOL foi pensada como oposto conceitual ao **Lua** (*Lua* = lua, *Sol* = sol): mesma ideia de linguagem de script, mas com classes, tipos explícitos e encapsulamento.

---

## Palavras reservadas

### Keywords com tema solar

| Keyword | Uso | Equivalente familiar |
|---------|-----|----------------------|
| `shine` | Declarar classe | `class` |
| `glow` | Construtor | `constructor` |
| `ray` | Declarar método | método |
| `eclipse` | Herança ou chamada ao pai | `extends` / `super` |
| `emit` | Retornar valor de um método | `return` |
| `flare` | Lançar exceção | `throw` |

### Keywords gerais

```
public   private   var   new   this
if       else      while
for      each      in
try      catch
true     false     null
int      float     bool   string   void
```

---

## Tipos

Todo campo, parâmetro e variável deve ter tipo explícito.

| Tipo | Descrição | Exemplo |
|------|-----------|---------|
| `int` | Inteiro 64-bit | `var n int = 42;` |
| `float` | Ponto flutuante 64-bit | `var x float = 3.14;` |
| `bool` | Booleano | `var ok bool = true;` |
| `string` | Texto | `var msg string = "olá";` |
| `void` | Sem valor (métodos sem retorno) | omitir `emit` |
| *NomeDeClasse* | Referência a objeto | `var c ContaBancaria = ...;` |
| `[Tipo]` | Array | `var nums [int] = [1, 2, 3];` |

### Declaração de variável

```sol
var nome Tipo;
var nome Tipo = expressao;
```

```sol
var saldo float = 1000.00;
var titular string = "Ana Silva";
var conta ContaBancaria = new ContaBancaria("Ana", 1000.00);
```

---

## Literais e comentários

### Literais

```sol
42          // int
3.14        // float
"texto"     // string (aspas duplas)
true        // bool
false       // bool
null        // ausência de valor
```

### Comentários

```sol
// comentário de linha

/* comentário
   de bloco */
```

---

## Classes e OOP

### Declarar classe (`shine`)

```sol
shine NomeDaClasse {

    private float saldo;
    public string titular;

    glow(string titular, float saldoInicial) {
        this.titular = titular;
        this.saldo   = saldoInicial;
    }

    public ray depositar(float valor) {
        this.saldo = this.saldo + valor;
    }

    public ray getSaldo() float {
        emit this.saldo;
    }
}
```

| Elemento | Sintaxe |
|----------|---------|
| Campo | `(public \| private) Tipo nome;` |
| Construtor | `glow(Tipo param, ...) { ... }` |
| Método | `(public \| private) ray nome(Tipo param, ...) [TipoRetorno] { ... }` |
| Retorno | `emit expressao;` ou `emit;` (void) |
| Instanciar | `new NomeDaClasse(arg1, arg2, ...)` |
| Acesso ao objeto atual | `this.campo` |
| Chamada de método | `obj.metodo(arg1, arg2)` |

### Herança (`eclipse`)

```sol
shine Filha eclipse Pai {

    glow(string nome) {
        eclipse.glow(nome, 0.0);   // chama construtor do pai
    }

    public ray sacar(float valor) {
        eclipse.sacar(valor);      // chama método do pai
    }
}
```

- Na declaração da classe: `shine Filha eclipse Pai` — herda de `Pai`.
- No corpo do método: `eclipse.metodo(...)` — equivalente ao `super` do Java.

---

## Variáveis e atribuição

```sol
var x int = 10;
x = x + 1;
this.saldo = this.saldo - valor;
```

- Campos de objeto: `this.nomeDoCampo`
- Variáveis locais dentro de métodos/blocos: `var nome Tipo = valor;`
- Atribuição a campo: `this.campo = expressao;`

---

## Operadores

### Aritméticos

| Operador | Operação |
|----------|----------|
| `+` | Adição |
| `-` | Subtração |
| `*` | Multiplicação |
| `/` | Divisão |
| `%` | Resto (módulo) |

### Comparação

| Operador | Operação |
|----------|----------|
| `==` | Igual |
| `!=` | Diferente |
| `<` | Menor |
| `>` | Maior |
| `<=` | Menor ou igual |
| `>=` | Maior ou igual |

### Lógicos

| Operador | Operação |
|----------|----------|
| `&&` | E lógico |
| `\|\|` | Ou lógico |
| `!` | Negação |

### Precedência (maior → menor)

1. `!`, `-` (unário)
2. `*`, `/`, `%`
3. `+`, `-`
4. `<`, `>`, `<=`, `>=`
5. `==`, `!=`
6. `&&`
7. `\|\|`

Use parênteses para forçar ordem: `(a + b) * c`.

---

## Controle de fluxo

### `if` / `else`

```sol
if (condicao) {
    // then
} else {
    // else
}
```

A condição deve ser do tipo `bool` (ou expressão comparada).

### `while`

```sol
var i int = 0;
while (i < 10) {
    i = i + 1;
}
```

### `for each` (arrays)

```sol
var nums [int] = [1, 2, 3];
var first int = nums[0];
var n int = nums.length;

for each x in nums {
    Console.print("x=" + x);
}
```

A expressão após `in` deve ser um array (`[Tipo]`). Arrays suportam indexação `arr[i]` (índice `int`) e a propriedade `.length`.

---

## Stdlib: `Console`

A classe **`Console`** faz parte da biblioteca padrão e está sempre disponível (não precisa ser declarada). É a forma recomendada de escrever saída no console:

```sol
Console.print("Hello, SOL");
Console.print("saldo=", conta.getSaldo());
Console.print("n=" + 42);
```

`Console.print` aceita um ou mais argumentos dos tipos `string`, `int`, `float` ou `bool`. Vários argumentos são separados por espaço na saída.

---

## Concatenação de strings

O operador `+` concatena strings. Valores numéricos e booleanos são convertidos para texto automaticamente:

```sol
var msg string = "saldo=" + 1150;
var linha string = "ok=" + true;
Console.print(msg);
```

`string + string`, `string + int`, `string + float` e `string + bool` produzem `string`. Combinações inválidas (ex.: `string + ContaBancaria`) geram erro em compile-time.

---

## Exceções (`flare` / `try` / `catch`)

### Lançar exceção

```sol
flare "Mensagem de erro";
flare expressaoString;
```

`flare` só é permitido **dentro de métodos** (não no top-level do script).

### Capturar exceção

```sol
try {
    conta.sacar(9999.00);
} catch (erro) {
    var msg string = erro;   // erro contém a mensagem (string)
}
```

Se `flare` não for capturado, a VM aborta a execução com erro.

---

## Programa script (top-level)

Um programa `.sol` é uma sequência de declarações de classe e statements no nível raiz:

```sol
shine MinhaClasse {
    // ...
}

var obj MinhaClasse = new MinhaClasse();
obj.metodo();
```

Não existe `main()`. O compilador gera um bloco `__program` que executa os statements top-level após carregar as classes.

---

## Como executar

```bash
# Compilar (gerar TAC)
./solc --compile programa.sol
./solc --compile programa.sol -o saida.tac

# Executar (compilar + interpretar)
./solc --run programa.sol

# Verificar tipos sem executar
./solc --check programa.sol

# Inspecionar fases do compilador
./solc --lex programa.sol      # tokens
./solc --parse programa.sol    # AST (resumo)
```

Exemplos incluídos no repositório:

```bash
./solc --run examples/hello.sol
./solc --run examples/script.sol
./solc --run examples/conta_bancaria.sol
./solc --run examples/heranca.sol
./solc --run examples/for_each.sol
./solc --run examples/simple.sol
```

---

## Logs e saída no console

Use `Console.print(...)` para escrever no stdout ao executar com `./solc --run`:

```bash
./solc --run examples/hello.sol
# Hello, SOL
# program finished OK
```

| Objetivo | Comando / técnica |
|----------|-------------------|
| Executar script com saída | `./solc --run arquivo.sol` |
| Binário nativo | `./solc --build arquivo.sol -o programa && ./programa` |
| Ver tokens | `./solc --lex arquivo.sol` |
| Ver erros de tipo | `./solc --check arquivo.sol` |
| Ver código intermediário | `./solc --compile arquivo.sol` → `output.tac` |
| Erros em runtime | `flare` não capturado → stderr: `solc: flare: mensagem` |

Exemplos prontos:

```bash
./solc --run examples/hello.sol
./solc --run examples/for_each.sol
./solc --run examples/script.sol
./solc --run examples/conta_bancaria.sol
```

---

## Limitações atuais

| Recurso | Status |
|---------|--------|
| Lexer, parser, semântica | Implementado |
| Execução via VM (`--run`) | Implementado |
| OOP, herança, `flare`, `try/catch` | Implementado |
| `Console.print` (stdlib) | Implementado |
| Concatenação de strings (`+`) | Implementado |
| Arrays, `for each`, `arr[i]`, `.length` | Implementado |
| Backend nativo (`--build`) | Experimental (hello world + subset) |
| Múltipla herança | Não suportado |
| Modificadores além de public/private | Não suportado |
| Tipos genéricos | Não suportado |

> Plano detalhado para resolver cada item acima: [`PLANO-LIMITACOES.md`](PLANO-LIMITACOES.md)

---

## Exemplos completos

### Conta bancária

```sol
shine ContaBancaria {

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

    public ray getSaldo() float {
        emit this.saldo;
    }
}

var conta ContaBancaria = new ContaBancaria("Ana Silva", 1000.00);
conta.depositar(250.00);
conta.sacar(100.00);
// saldo final: 1150.00
```

### Herança

```sol
shine ContaEspecial eclipse ContaBancaria {

    private float limiteCredito;

    glow(string titular, float saldoInicial, float limiteCredito) {
        eclipse.glow(titular, saldoInicial);
        this.limiteCredito = limiteCredito;
    }

    public ray sacar(float valor) {
        var disponivel float = this.getSaldo() + this.limiteCredito;
        if (valor > disponivel) {
            flare "Total limit exceeded";
        }
        eclipse.sacar(valor);
    }
}
```

### Tratamento de erro

```sol
try {
    conta.sacar(99999.00);
} catch (erro) {
    var mensagem string = erro;
}
```

---

## Referências no repositório

| Arquivo | Conteúdo |
|---------|----------|
| [`SYNTAX.md`](SYNTAX.md) | Referência de sintaxe e limitações |
| [`PLANO-LIMITACOES.md`](PLANO-LIMITACOES.md) | Plano para implementar as limitações pendentes |
| [`README.md`](README.md) | Build e uso do `solc` |
| [`examples/`](examples/) | Programas `.sol` de exemplo |

---

_Versão do guia: 1.1 — VM interpretada (`solc --run`) com `Console.print`, arrays completos e backend nativo experimental._
