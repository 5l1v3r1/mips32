package main

import (
	"encoding/hex"
	"errors"
	"strconv"
	"strings"

	"github.com/gopherjs/gopherjs/js"
	"github.com/unixpickle/mips32"
)

func DisassembleData() {
	textarea := js.Global.Get("disassembler-data")
	text := textarea.Get("value").String()
	text = strings.Replace(text, " ", "", -1)
	text = strings.Replace(text, "\n", "", -1)
	text = strings.Replace(text, "\t", "", -1)

	data, err := hex.DecodeString(text)
	if err != nil {
		showDisassemblerError(err)
		return
	} else if len(data)%4 != 0 {
		showDisassemblerError(errors.New("binary data's size must be a multiple of 32-bits"))
		return
	}

	instructions := make([]*mips32.Instruction, len(data)/4)
	instStrs := make([]string, len(instructions))
	for i := 0; i < len(data); i += 4 {
		word := (uint32(data[i+0]) << 24) | (uint32(data[i+1]) << 16) | (uint32(data[i+2]) << 8) |
			uint32(data[i+3])
		instructions[i/4] = mips32.DecodeInstruction(word)
		rendering, err := instructions[i/4].Render()
		if err != nil {
			showDisassemblerError(err)
			return
		}
		instStrs[i/4] = rendering.String()
	}

	hideDisassemblerError()
	SetAssemblerCode(strings.Join(instStrs, "\n"))
	ShowAssembler()
}

func SetDisassemblerData(data []uint32) {
	textarea := js.Global.Get("assembler-code")

	spacedValues := make([]string, len(data))
	for i, d := range data {
		numStr := strconv.FormatUint(uint64(d), 16)
		for len(numStr) < 8 {
			numStr = "0" + numStr
		}
		spacedValues[i] = numStr
	}

	textarea.Set("value", strings.Join(spacedValues, " "))
}

func hideDisassemblerError() {
	errField := js.Global.Get("disassembler-error")
	errField.Set("className", "error-view")
}

func showDisassemblerError(err error) {
	errField := js.Global.Get("disassembler-error")
	errField.Set("className", "error-view showing-error")
	errField.Set("innerText", err.Error())
}
