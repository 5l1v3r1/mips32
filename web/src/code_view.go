package main

import (
	"github.com/gopherjs/gopherjs/js"
	"github.com/unixpickle/mips32"
)

const PreviewLineCount = 7

type CodeView struct {
	element *js.Object
}

func NewCodeView() *CodeView {
	return &CodeView{
		element: js.Global.Get("debugger-code-view"),
	}
}

func (c *CodeView) Update(e *mips32.Emulator) {
	var startAddress uint32
	if (e.ProgramCounter / 4) > (PreviewLineCount / 2) {
		startAddress = e.ProgramCounter - (PreviewLineCount/2)*4
	}
	c.element.Set("innerHTML", "")
	for i := 0; i < PreviewLineCount; i++ {
		addr := startAddress + uint32(i*4)
		row := createCodeViewLine(e, addr)
		if addr == e.ProgramCounter {
			row.Set("className", row.Get("className").String()+" debugger-code-view-current")
		}
		c.element.Call("appendChild", row)
	}
}

func createCodeViewLine(e *mips32.Emulator, addr uint32) *js.Object {
	document := js.Global.Get("document")
	row := document.Call("createElement", "tr")

	addrColumn := document.Call("createElement", "td")
	addrColumn.Set("innerText", format32BitHex(addr))
	addrColumn.Set("className", "debugger-code-view-addr")
	row.Call("appendChild", addrColumn)

	codeColumn := document.Call("createElement", "td")
	codeColumn.Set("className", "debugger-code-view-code")
	if inst := e.Executable.Get(addr); inst != nil {
		rendering, err := inst.Render()
		if err != nil {
			codeColumn.Set("innerText", "(Unknown)")
		} else {
			codeColumn.Set("innerText", rendering.String())
		}
	} else {
		codeColumn.Set("innerText", "NOP")
	}
	row.Call("appendChild", codeColumn)

	opcodeColumn := document.Call("createElement", "td")
	opcodeColumn.Set("className", "debugger-code-view-opcode")
	if inst := e.Executable.Get(addr); inst != nil {
		opcode, err := inst.Encode(addr, e.Executable.Symbols)
		if err != nil {
			codeColumn.Set("innerText", "(invalid)")
		} else {
			codeColumn.Set("innerText", format32BitHex(opcode))
		}
	} else {
		codeColumn.Set("innerText", format32BitHex(0))
	}

	return row
}
