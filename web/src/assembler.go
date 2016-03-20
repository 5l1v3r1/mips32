package main

import (
	"github.com/gopherjs/gopherjs/js"
	"github.com/unixpickle/mips32"
)

const maxAssembleSize = 0x1000

type Assembler struct {
	textarea  *js.Object
	errorView *js.Object
}

func NewAssembler() *Assembler {
	res := &Assembler{
		textarea:  js.Global.Get("assembler-code"),
		errorView: js.Global.Get("assembler-error"),
	}
	js.Global.Get("assembler-button").Call("addEventListener", "click", res.assemble)
	return res
}

func (a *Assembler) SetCode(code string) {
	a.textarea.Set("value", code)
	a.hideError()
}

func (a *Assembler) Show() {
	js.Global.Get("location").Set("hash", "#assembler")
}

func (a *Assembler) assemble() {
	text := a.textarea.Get("value").String()

	lines, err := mips32.TokenizeSource(text)
	if err != nil {
		a.showError(err)
		return
	}
	exc, err := mips32.ParseExecutable(lines)
	if err != nil {
		a.showError(err)
		return
	}
	a.hideError()
	GlobalDebugger.SetExecutable(exc)
	GlobalDebugger.Show()

	if exc.End() < maxAssembleSize {
		data := make([]uint32, exc.End()/4)
		for addr := uint32(0); addr < exc.End(); addr += 4 {
			if inst := exc.Get(addr); inst != nil {
				data[addr/4], _ = inst.Encode(addr, exc.Symbols)
			} else {
				data[addr/4] = 0
			}
		}
		GlobalDisassembler.SetData(data)
	}
}

func (a *Assembler) hideError() {
	a.errorView.Set("className", "error-view")
}

func (a *Assembler) showError(err error) {
	a.errorView.Set("className", "error-view showing-error")
	a.errorView.Set("innerText", err.Error())
}
