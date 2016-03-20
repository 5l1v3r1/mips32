package main

import "github.com/gopherjs/gopherjs/js"

func main() {
	js.Global.Get("window").Call("addEventListener", "load", func() {
		js.Global.Get("assembler-button").Call("addEventListener", "click", AssembleCode)
		js.Global.Get("disassembler-button").Call("addEventListener", "click", DisassembleData)
	})
}
