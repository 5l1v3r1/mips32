package main

import (
	"strconv"
	"sync"
	"time"

	"github.com/gopherjs/gopherjs/js"
	"github.com/unixpickle/mips32"
)

type Debugger struct {
	lock sync.Mutex

	frequency   int
	controlChan chan debuggerCommand
	emulator    *mips32.Emulator

	registers *Registers
	codeView  *CodeView
}

func NewDebugger() *Debugger {
	res := &Debugger{
		frequency:   4,
		controlChan: make(chan debuggerCommand, 0),
		emulator: &mips32.Emulator{
			Memory: mips32.NewLazyMemory(),
			Executable: &mips32.Executable{
				Segments: map[uint32][]mips32.Instruction{},
				Symbols:  map[string]uint32{},
			},
			LittleEndian: true,
		},
		registers: NewRegisters(),
		codeView:  NewCodeView(),
	}

	go res.debugLoop()

	res.registers.SetCallback(func(reg int, val uint32) {
		res.lock.Lock()
		defer res.lock.Unlock()
		res.emulator.RegisterFile[reg] = val
	})

	res.registerUIEvents()
	res.updateUI()

	return res
}

func (d *Debugger) SetExecutable(e *mips32.Executable) {
	d.controlChan <- stopDebugger
	d.lock.Lock()
	d.emulator = &mips32.Emulator{
		Memory:       mips32.NewLazyMemory(),
		Executable:   e,
		LittleEndian: true,
	}
	d.lock.Unlock()
	d.updateUI()
}

func (d *Debugger) debugLoop() {
	for command := range d.controlChan {
		if command == stepDebugger {
			d.stepDebugger()
		} else if command == startDebugger {
			d.runDebugger()
		}
	}
}

func (d *Debugger) runDebugger() {
	frameDuration, cyclesPerFrame := d.tickInfo()
	ticker := time.NewTicker(frameDuration)
	defer func() {
		ticker.Stop()
	}()
	defer d.updateButtonState(false)
	d.updateButtonState(true)
	for {
		select {
		case <-ticker.C:
		case msg := <-d.controlChan:
			if msg == stopDebugger {
				return
			} else if msg == updateDebuggerFreq {
				ticker.Stop()
				frameDuration, cyclesPerFrame = d.tickInfo()
				ticker = time.NewTicker(frameDuration)
			}
		}
		d.lock.Lock()
		for i := 0; i < cyclesPerFrame; i++ {
			err := d.emulator.Step()
			if err != nil {
				d.lock.Unlock()
				d.handleError(err)
				return
			}
			if d.emulator.Done() {
				d.lock.Unlock()
				d.updateUI()
				return
			}
		}
		d.lock.Unlock()
		d.updateUI()
	}
}

func (d *Debugger) stepDebugger() {
	d.lock.Lock()
	err := d.emulator.Step()
	d.lock.Unlock()
	if err != nil {
		d.handleError(err)
	} else {
		d.updateUI()
	}
}

func (d *Debugger) handleError(err error) {
	// TODO: show error here.
	d.updateButtonState(false)
	println("error:", err.Error())
}

func (d *Debugger) updateUI() {
	d.lock.Lock()
	defer d.lock.Unlock()

	d.registers.Update(d.emulator.RegisterFile)
	d.codeView.Update(d.emulator)
}

func (d *Debugger) updateButtonState(running bool) {
	d.lock.Lock()
	defer d.lock.Unlock()

	button := js.Global.Get("debugger-play")
	if running {
		button.Set("innerText", "Stop")
	} else {
		button.Set("innerText", "Play")
	}
}

func (d *Debugger) tickInfo() (framePause time.Duration, instsPerFrame int) {
	d.lock.Lock()
	defer d.lock.Unlock()
	if d.frequency > 32 {
		return time.Second / 32, d.frequency / 32
	} else {
		return time.Second / time.Duration(d.frequency), 1
	}
}

func (d *Debugger) registerUIEvents() {
	js.Global.Get("debugger-step").Call("addEventListener", "click", func() {
		d.controlChan <- stepDebugger
	})

	playButton := js.Global.Get("debugger-play")
	playButton.Call("addEventListener", "click", func() {
		if playButton.Get("innerText").String() == "Stop" {
			d.controlChan <- stopDebugger
		} else {
			d.controlChan <- startDebugger
		}
	})

	ratePicker := js.Global.Get("debugger-rate")
	ratePicker.Call("addEventListener", "change", func() {
		rate := ratePicker.Get("value").String()
		num, _ := strconv.Atoi(rate)
		d.lock.Lock()
		d.frequency = num
		d.lock.Unlock()
		d.controlChan <- updateDebuggerFreq
	})

	js.Global.Get("debugger-reset").Call("addEventListener", "click", func() {
		d.controlChan <- stopDebugger
		d.lock.Lock()
		d.emulator = &mips32.Emulator{
			Memory:       mips32.NewLazyMemory(),
			Executable:   d.emulator.Executable,
			LittleEndian: true,
		}
		d.lock.Unlock()
		d.updateUI()
	})
}

type debuggerCommand int

const (
	stopDebugger debuggerCommand = iota
	startDebugger
	stepDebugger
	updateDebuggerFreq
)
