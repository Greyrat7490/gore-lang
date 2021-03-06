# gore-lang

> Fun Project to write my own language, to learn Go and to see how a good language would look like for me. So it will probably end up like a modern C ... so Go

A statically and strongly typed programming language similar to Go, but with more focus on memory. It is more like a mix of C, Go and Rust.
It is named Gore, because I took everything out of C, Go and Rust I like to write this bloody mess.
And because it is written in Go...

* fast
* easy
* compiled
* statically and strongly typed
* lightweigth
* important build-in functions
* designed around hardware-near programming
* crossplatform

## Supported:
* [x] Linux
* [ ] MacOS
* [ ] windows
* [x] x86_64
* [ ] ARM

## TODO:
* [x] generate assembly file
  * [x] nasm
  * [ ] fasm (preferable!)
* [x] generate executable
* [x] variables
* [x] functions
* [x] syscalls
* [x] arithmetics
  * [x] unary ops
  * [x] binary ops
    * [x] parse by precedence
  * [x] parentheses
* [x] controll structures
  * [x] if
  * [x] else
  * [x] elif
  * [x] while
  * [x] for
  * [x] switch
  * [x] xswitch (expr switch)
  * [x] &&, ||
* [x] pointer
  * [x] define/assign
  * [x] deref
  * [x] get addr (via "&")
  * [x] arithmetic
* [x] consts
  * [x] define
  * [x] compile time eval
* [ ] arrays
  * [ ] declare
    * [ ] with literal
    * [ ] with const
* [ ] structs
* [x] turing complete -> actual programming language
  * [x] proof with Rule 110 programm
* [x] type checking
* [x] tests
* [ ] examples
* [ ] self-hosted
* [ ] cross-platform

## Get Started

compile a source file
```console
$ go run gorec <source_file>
```
run tests
```console
$ go test ./test -v
```
gorec usage
```console
$ go run gorec --help
gorec usage:
  -ast
    	show the AST
  -r	run the compiled executable
```
