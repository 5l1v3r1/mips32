package main

import (
	"github.com/gopherjs/gopherjs/js"
	"github.com/unixpickle/mips32"
)

func main() {
	js.Global.Get("window").Call("addEventListener", "load", func() {
		js.Global.Get("assembler-button").Call("addEventListener", "click", AssembleCode)
		js.Global.Get("disassembler-button").Call("addEventListener", "click", DisassembleData)

		r := NewRegisters()
		r.Update(mips32.RegisterFile{13: 0x1337, 21: 0x13381337})
	})
}
