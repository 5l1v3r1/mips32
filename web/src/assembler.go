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
	// TODO: something.
	println(exc)
}

func hideAssemblerError() {
	errField := js.Global.Get("assembler-error")
	errField.Set("className", "")
}

func showAssemblerError(err error) {
	errField := js.Global.Get("assembler-error")
	errField.Set("className", "showing-error")
	errField.Set("innerText", err.Error())
}
