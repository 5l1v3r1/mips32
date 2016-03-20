package main

import (
	"github.com/gopherjs/gopherjs/js"
	"github.com/unixpickle/mips32"
)

func AssembleCode() {
	textarea := js.Global.Get("assembler-code")
	text := textarea.Get("value").String()

	lines, err := mips32.TokenizeSource(text)
	if err != nil {
		showAssemblerError(err)
		return
	}
	exc, err := mips32.ParseExecutable(lines)
	if err != nil {
		showAssemblerError(err)
		return
	}
	hideAssemblerError()
	GlobalDebugger.SetExecutable(exc)
	GlobalDebugger.Show()
}

func SetAssemblerCode(code string) {
	textarea := js.Global.Get("assembler-code")
	textarea.Set("value", code)
	hideAssemblerError()
}

func ShowAssembler() {
	js.Global.Get("location").Set("hash", "#assembler")
}

func hideAssemblerError() {
	errField := js.Global.Get("assembler-error")
	errField.Set("className", "error-view")
}

func showAssemblerError(err error) {
	errField := js.Global.Get("assembler-error")
	errField.Set("className", "error-view showing-error")
	errField.Set("innerText", err.Error())
}
