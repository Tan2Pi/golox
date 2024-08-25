# GoLox

glox is an interpreter written in Go, for the Lox programming language used in the book [crafting interpreters](https://craftinginterpreters.com/).
It is a relatively faithful port of the Java tree-walk interpeter (jlox) from the book.

## Install

`glox` can be installed using the Go toolchain. The minimum Go version is 1.23.
```bash
go install github.com/Tan2Pi/golox@latest
```

## Usage

### REPL

```bash
glox
```

### Source Code

```bash
glox $FILE
```

## Notable Changes and Features

## Collections

glox has simple builtin List and Map types, backed by native Go slices & maps.
For simplicity in the implementationn, this is done by leveraging Lox _class_ functionality,
rather than adding direct language support for them. As a result, there's no subscript operator
like `data[1]`, as you might expect in a real programming language. In addition, there's no way
to iterate over a map with this approach.

A basic example usage of glox's collections:

```js
// maps
var map = Map();
map.put("hello", "world!");
map.put("one", "two");
print map.contains("two");
print map.get("one");

// lists
var list = List();
list.append(1);
list.append(2);
print list;
```

### Exceptions

jlox uses exceptions in situations where unwinding the call stack is required, such as return statements.
To replicate thie, the panic/recover pattern was initially used, but while benchmarking I noticed that
panic/recover was slower than expected. As a result, instead of panic/recover, glox instead propagates return
statement results by manually propagating them through the call stack using simple returns.

## Building & Testing

`go build ./...` and `go test ./...` is pretty much all you need.

The test suite used is a simplified port of the test suite from the book's
[github repository](https://github.com/munificent/craftinginterpreters?tab=readme-ov-file#testing-your-implementation).
The test cases themselves are lox code with inlined assertions, which are unchanged from the source repository.

This was done primarily to keep testing self-contained and runnable with `go test`.

## Changelog
* 05/15/2022:
    - Finished Chapter 6: Parsing Expressions
    - Bug: Tokens with TokenType Number have integer Literal values.
      These should be floating point numbers, given that Lox only has one numerical type.
    - Need to verify if AstPrinter is rendering correctly or if there's a parsing bug
    - Verify that errors act as they should. Using panic() and recover() instead of
      propagating errors up may not be the best way.
* 05/16/2022:
    - Halfway through Chapter 7: Evaluating Expressions
    - Fixed bug reg: TokenType Number
    - Probably easier to verify functionality if the Interpreter is wired up
* 011/5/2023 - The rest of the owl:
    - Complete implementation of Golox
    - Add GH actions for presubmit tests.
* 08/25/2024 - Odds and ends:
    - Fix unicode comments (finally)
    - Port test suite to Go to simplify testing and CI
    - Add List and Map classes for simple collections
    - Add golangci-lint
