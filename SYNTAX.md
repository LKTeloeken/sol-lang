# SOL — Referência de Sintaxe

Guia prático da linguagem **SOL** (*Scripting Object Language*): tipos, keywords, operadores, OOP e como usar o interpretador.

> SOL é estaticamente tipada, orientada a objetos e executada como **script** (sem função `main`). O interpretador `sollang` analisa o código e uma VM interpreta o TAC gerado.

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
| Interpretador | `sollang` |

SOL foi pensada como oposto conceitual ao **Lua** (*Lua* = lua, *Sol* = sol): mesma ideia de linguagem de script, mas com classes, tipos explícitos e encapsulamento.

---

## Palavras reservadas

### Keywords com tema solar

| Keyword | Uso | Equivalente familiar |
|---------|-----|----------------------|
| `rise` | Declarar classe | `class` |
| `glow` | Construtor | `constructor` |
| `ray` | Declarar método | método |
| `radiate` | Herança ou chamada ao pai | `extends` / `super` |
| `emit` | Retornar valor de um método | `return` |
| `flare` | Lançar exceção | `throw` |
| `orbit` | Importar outro arquivo `.sol` | `import` / `#include` |
| `star` | Alias de tipo reutilizável | `type` / `typedef` |

### Keywords gerais

```
public   private   var   new   this
if       else      while
for      each      in
break    continue
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
| `Tipo[]` | Array (sintaxe postfix) | `var nums int[] = [1, 2, 3];` |

### Alias de tipo (`star`)

Declare um nome reutilizável para um tipo (top-level):

```sol
star TodoItems = string[];
star Contas = ContaBancaria[];

rise TodoList {
    private TodoItems items;
}
```

Equivalente semanticamente a usar `string[]` ou `ContaBancaria[]` diretamente.

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

### Declarar classe (`rise`)

```sol
rise NomeDaClasse {

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

### Herança (`radiate`)

```sol
rise Filha radiate Pai {

    glow(string nome) {
        radiate.glow(nome, 0.0);   // chama construtor do pai
    }

    public ray sacar(float valor) {
        radiate.sacar(valor);      // chama método do pai
    }
}
```

- Na declaração da classe: `rise Filha radiate Pai` — herda de `Pai`.
- No corpo do método: `radiate.metodo(...)` — equivalente ao `super` do Java.

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

Loop infinito (estilo Go):

```sol
while (true) {
    // corpo
}
```

### `for i in 0..10` (range inteiro)

```sol
for i in 0..10 {
    Console.print(i);
}
```

Itera com `i` de `0` até `9` (intervalo semi-aberto: inclui o início, exclui o fim). Os limites podem ser expressões `int`:

```sol
var start int = 2;
var end int = 5;
for i in start..end {
    Console.print(i);
}
```

### `for each` (arrays)

```sol
var nums int[] = [1, 2, 3];
var first int = nums[0];
var n int = nums.length;

for each x in nums {
    Console.print("x=" + x);
}
```

A expressão após `in` deve ser um array (`Tipo[]`). Arrays suportam indexação `arr[i]` (índice `int`), a propriedade `.length` e métodos de instância:

| Método | Assinatura | Efeito |
|--------|------------|--------|
| `push` | `(elem T) void` | adiciona ao final |
| `pop` | `() T` | remove e retorna o último |
| `remove` | `(index int) void` | remove por índice |
| `insert` | `(index int, elem T) void` | insere na posição |
| `contains` | `(elem T) bool` | busca linear |
| `clear` | `() void` | esvazia o array |
| `isEmpty` | `() bool` | `true` se vazio |

```sol
var items string[] = [];
items.push("a");
items.remove(0);
```

Literal vazio `[]` exige tipo declarado no LHS: `var xs string[] = [];`

### `break` e `continue`

```sol
for i in 0..100 {
    if (i == 50) {
        break;
    }
    if (i % 2 == 0) {
        continue;
    }
    Console.print(i);
}
```

Só podem ser usados dentro de loops (`while`, `for each`, `for i in ..`).

---

## Stdlib: `Console`

A classe **`Console`** faz parte da biblioteca padrão e está sempre disponível (não precisa ser declarada). É a forma recomendada de escrever saída no console:

```sol
Console.print("Hello, SOL");
Console.print("saldo=", conta.getSaldo());
Console.print("n=" + 42);
```

`Console.print` aceita um ou mais argumentos dos tipos `string`, `int`, `float` ou `bool`. Vários argumentos são separados por espaço na saída.

### Entrada do usuário

```sol
var nome string = Console.readLine();
var idade string = Console.readLine("Digite sua idade: ");
var n int = Console.readInt();
```

- `readLine()` — lê uma linha de texto do stdin (sem prompt).
- `readLine(prompt string)` — imprime o prompt e lê a linha.
- `readInt()` — lê uma linha e converte para `int` (erro em runtime se inválido).

---

## Stdlib: `File`

A classe **`File`** fornece leitura e escrita de arquivos inteiros como `string`:

```sol
var conteudo string = File.read("dados.txt");
File.write("saida.txt", conteudo);
File.write("log.txt", "linha\n");
```

- `File.read(path string) string` — lê o arquivo completo; erro em runtime se o arquivo não existir.
- `File.write(path string, content string) void` — sobrescreve o arquivo com o conteúdo.
- `File.append(path string, content string) void` — acrescenta ao final do arquivo (cria se não existir).
- `File.exists(path string) bool` — `true` se o caminho existir.

---

## Stdlib: `Time`

```sol
var now int = Time.now();
Time.sleepMillis(100);
var s string = Time.format(now, "2006-01-02 15:04:05");
```

| Método | Assinatura | Descrição |
|--------|------------|-----------|
| `now` | `() int` | Unix timestamp (segundos) |
| `sleepMillis` | `(ms int) void` | Pausa a execução |
| `format` | `(unix int, layout string) string` | Formata data/hora (layouts como em Go; ver `time.Format`) |

---

## Stdlib: `String`

```sol
var n int = String.length("abc");
var t string = String.trim("  hi  ");
var parts string[] = String.split("a,b", ",");
var ok bool = String.contains("abc", "b");
var sub string = String.substring("hello", 1, 4);
```

| Método | Assinatura |
|--------|------------|
| `length` | `(s string) int` |
| `trim` | `(s string) string` |
| `split` | `(s string, sep string) string[]` |
| `contains` | `(s string, sub string) bool` |
| `substring` | `(s string, start int, end int) string` |

`substring` falha em runtime se os índices forem inválidos.

---

## Stdlib: `Math`

```sol
var a float = Math.abs(-1.5);
var lo float = Math.min(1.0, 2.0);
var hi float = Math.max(1.0, 2.0);
var n int = Math.floor(3.9);
var r float = Math.random();
```

| Método | Assinatura |
|--------|------------|
| `abs` | `(x float) float` |
| `min` / `max` | `(a float, b float) float` |
| `floor` | `(x float) int` |
| `random` | `() float` | valor em `[0, 1)` |

---

## Stdlib: `Args` (argumentos do script)

Argumentos passados ao `sollang` **depois** do arquivo `.sol`:

```bash
./sollang script.sol arg1 arg2
```

```sol
var n int = Args.count();
var first string = Args.at(0);
```

| Método | Assinatura |
|--------|------------|
| `count` | `() int` |
| `at` | `(i int) string` | erro (`flare`) se índice fora do intervalo |

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
rise MinhaClasse {
    // ...
}

var obj MinhaClasse = new MinhaClasse();
obj.metodo();
```

Não existe `main()`. O gerador de TAC cria um bloco `__program` que executa os statements top-level após carregar as classes.

### Imports (`orbit`)

Use `orbit` para incluir outro arquivo `.sol` no programa. O caminho é relativo ao diretório do arquivo que contém o `orbit`.

```sol
orbit "utils.sol";
```

- Mescla **classes e statements** top-level do arquivo importado na posição do `orbit`.
- Imports aninhados são suportados (`utils.sol` pode importar outros arquivos).
- Import circular é erro de compilação.
- `-parse` mostra o nó `orbit` no AST; `-check`, `-tac` e a execução expandem os imports antes da análise semântica.

Exemplo em `examples/modules/`:

```bash
./sollang examples/modules/main.sol
```

---

## Como executar

```bash
# Executar (padrão: compila para TAC e interpreta)
./sollang programa.sol

# Verificar tipos sem executar
./sollang -check programa.sol

# Inspecionar fases do pipeline
./sollang -lex programa.sol      # tokens
./sollang -parse programa.sol    # AST (resumo)
./sollang -tac programa.sol         # código intermediário (TAC) no stdout
./sollang -tac -o saida.tac programa.sol   # TAC em arquivo (-o antes do .sol)
```

Exemplos incluídos no repositório:

```bash
./sollang examples/hello.sol
./sollang examples/script.sol
./sollang examples/conta_bancaria.sol
./sollang examples/heranca.sol
./sollang examples/for_each.sol
./sollang examples/for_range.sol
./sollang examples/simple.sol
./sollang examples/modules/main.sol
```

---

## Logs e saída no console

Use `Console.print(...)` para escrever no stdout ao executar:

```bash
./sollang examples/hello.sol
# Hello, SOL
```

| Objetivo | Comando / técnica |
|----------|-------------------|
| Executar script com saída | `./sollang arquivo.sol` |
| Ver tokens | `./sollang -lex arquivo.sol` |
| Ver erros de tipo | `./sollang -check arquivo.sol` |
| Ver código intermediário (TAC) | `./sollang -tac arquivo.sol` |
| Erros em runtime | `flare` não capturado → stderr: `sollang: flare: mensagem` |

Exemplos prontos:

```bash
./sollang examples/hello.sol
./sollang examples/for_each.sol
./sollang examples/script.sol
./sollang examples/conta_bancaria.sol
```

---

## Limitações atuais

| Recurso | Status |
|---------|--------|
| Lexer, parser, semântica | Implementado |
| Execução via VM (padrão) | Implementado |
| OOP, herança, `flare`, `try/catch` | Implementado |
| `Console.print` (stdlib) | Implementado |
| `Console.readLine`, `Console.readInt` | Implementado |
| `File.read`, `File.write`, `File.append`, `File.exists` | Implementado |
| `Time.now`, `Time.sleepMillis`, `Time.format` | Implementado |
| `String.length`, `trim`, `split`, `contains`, `substring` | Implementado (VM) |
| `Math.abs`, `min`, `max`, `floor`, `random` | Implementado |
| `Args.count`, `Args.at` | Implementado |
| `for i in 0..10` (range) | Implementado |
| `break`, `continue` | Implementado |
| Concatenação de strings (`+`) | Implementado |
| Arrays postfix `Tipo[]`, `for each`, `arr[i]`, `.length`, métodos `push`/`pop`/… | Implementado (VM) |
| Aliases de tipo `star Nome = Tipo;` | Implementado |
| Imports `orbit "arquivo.sol"` | Implementado |
| Múltipla herança | Não suportado |
| Modificadores além de public/private | Não suportado |
| Tipos genéricos | Não suportado |

---

## Exemplos completos

### Conta bancária

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
rise ContaEspecial radiate ContaBancaria {

    private float limiteCredito;

    glow(string titular, float saldoInicial, float limiteCredito) {
        radiate.glow(titular, saldoInicial);
        this.limiteCredito = limiteCredito;
    }

    public ray sacar(float valor) {
        var disponivel float = this.getSaldo() + this.limiteCredito;
        if (valor > disponivel) {
            flare "Total limit exceeded";
        }
        radiate.sacar(valor);
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
| [`README.md`](README.md) | Build e uso do `sollang` |
| [`examples/`](examples/) | Programas `.sol` de exemplo |

---

_Versão do guia: 1.4 — Arrays postfix `T[]`, aliases `star`, métodos de lista na VM._
