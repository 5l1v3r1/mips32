package main

import (
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
		controlChan: make(chan debuggerCommand, 1),
		emulator:    nil,
		registers:   NewRegisters(),
		codeView:    NewCodeView(),
	}
	go res.debugLoop()

	res.registers.SetCallback(func(reg int, val uint32) {
		res.lock.Lock()
		defer res.lock.Unlock()
		if res.emulator != nil {
			res.emulator.RegisterFile[reg] = val
		}
	})

	// TODO: register events from buttons, etc.

	res.updateUI()
	return res
}

func (d *Debugger) ChangeExecutable(e *mips32.Executable) {
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
	defer ticker.Stop()
	defer d.updateButtonState(false)
	for {
		select {
		case <-ticker.C:
		case msg := <-d.controlChan:
			if msg == stopDebugger {
				return
			}
		}
		d.lock.Lock()
		if d.emulator == nil {
			d.lock.Unlock()
			continue
		}
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
	if d.emulator == nil {
		d.lock.Unlock()
		return
	}
	err := d.emulator.Step()
	d.lock.Unlock()
	if err != nil {
		d.handleError(err)
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

	if d.emulator != nil {
		d.registers.Update(d.emulator.RegisterFile)
		d.codeView.Update(d.emulator)
	} else {
		d.registers.Update(mips32.RegisterFile{})
		d.codeView.Update(&mips32.Emulator{Executable: &mips32.Executable{}})
	}
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

type debuggerCommand int

const (
	stopDebugger debuggerCommand = iota
	startDebugger
	stepDebugger
)
