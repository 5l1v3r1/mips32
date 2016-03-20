package main

import "github.com/gopherjs/gopherjs/js"

var GlobalDebugger *Debugger
var GlobalAssembler *Assembler
var GlobalDisassembler *Disassembler

func main() {
	js.Global.Get("window").Call("addEventListener", "load", func() {
		GlobalDebugger = NewDebugger()
		GlobalAssembler = NewAssembler()
		GlobalDisassembler = NewDisassembler()
	})
}
