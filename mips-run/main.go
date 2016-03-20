package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/unixpickle/mips32"
)

func main() {
	var littleEndian bool
	flag.BoolVar(&littleEndian, "little", false, "use little endian memory")

	var relaxAlignment bool
	flag.BoolVar(&relaxAlignment, "misaligned", false, "allow misaligned memory access")

	flag.Parse()
	if len(flag.Args()) != 1 {
		dieUsage()
	}

	file := flag.Args()[0]
	contents, err := ioutil.ReadFile(file)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	tokens, err := mips32.TokenizeSource(string(contents))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	exc, err := mips32.ParseExecutable(tokens)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	emu := &mips32.Emulator{
		Memory:            mips32.NewLazyMemory(),
		Executable:        exc,
		LittleEndian:      littleEndian,
		ForceMemAlignment: !relaxAlignment,
	}
	for !emu.Done() {
		if err := emu.Step(); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}

	fmt.Println("Register file:")
	fmt.Println(emu.RegisterFile.String())
}

func dieUsage() {
	fmt.Fprintln(os.Stderr, "Usage:", os.Args[0], "[flags] <file.s>")
	flag.PrintDefaults()
	os.Exit(1)
}
