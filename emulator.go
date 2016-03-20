package mips32

import "errors"

type RegisterFile [32]uint32

type Emulator struct {
	RegisterFile   RegisterFile
	Memory         Memory
	Executable     *Executable
	ProgramCounter uint32

	DelaySlot  bool
	JumpTarget uint32
}

func (e *Emulator) Step() error {
	inst := e.Executable.Get(e.ProgramCounter)
	if e.DelaySlot {
		e.DelaySlot = false
		e.ProgramCounter = e.JumpTarget
	} else {
		e.ProgramCounter += 4
	}

	// If there is no instruction in the ROM, we assume it is a NOP.
	if inst == nil {
		return nil
	}

	switch inst.Name {
	case "BEQ", "BGEZ", "BGTZ", "BLEZ", "BLTZ", "BNE":
		return e.executeBranch(inst)
	case "J", "JR", "JAL", "JALR":
		return e.executeJump(inst)
	case "LB", "LBU", "LW", "SB", "SW":
		return e.executeMemory(inst)
	case "ADDU", "AND", "NOR", "OR", "SUBU", "XOR":
		e.executeRegisterArithmetic(inst)
	case "ADDIU", "ANDI", "ORI", "XORI":
		e.executeImmediateArithmetic(inst)
	case "LUI":
		e.executeLoadUpperImmediate(inst)
	case "SLT", "SLTI", "SLTIU", "SLTU":
		e.executeSetLessThan(inst)
	case "SLL", "SRL", "SRA":
		e.executeConstantShift(inst)
	case "SLLV", "SRLV", "SRAV":
		e.executeRegisterShift(inst)
	case "MOVN", "MOVZ":
		e.executeConditionalMove(inst)
	default:
		return errors.New("unknown instruction: " + inst.Name)
	}
	return nil
}

func (e *Emulator) executeBranch(inst *Instruction) error {
	// TODO: this.
	return nil
}

func (e *Emulator) executeJump(inst *Instruction) error {
	// TODO: this.
	return nil
}

func (e *Emulator) executeMemory(inst *Instruction) error {
	// TODO: this.
	return nil
}

func (e *Emulator) executeRegisterArithmetic(inst *Instruction) {
	if inst.Registers[0] == 0 {
		return
	}
	val1 := e.RegisterFile[inst.Registers[1]]
	val2 := e.RegisterFile[inst.Registers[2]]
	var result uint32
	switch inst.Name {
	case "ADDU":
		result = val1 + val2
	case "AND":
		result = val1 & val2
	case "OR":
		result = val1 | val2
		fallthrough
	case "NOR":
		result = ^result
	case "SUBU":
		result = val1 - val2
	case "XOR":
		result = val1 ^ val2
	}
	e.RegisterFile[inst.Registers[0]] = result
}

func (e *Emulator) executeImmediateArithmetic(inst *Instruction) {
	if inst.Registers[0] == 0 {
		return
	}

	val1 := e.RegisterFile[inst.Registers[1]]
	var val2 uint32
	if inst.Name == "ADDIU" {
		val2 = uint32(inst.SignedConstant16)
	} else {
		val2 = uint32(inst.UnsignedConstant16)
	}

	var result uint32
	switch inst.Name {
	case "ADDIU":
		result = val1 + val2
	case "ANDI":
		result = val1 & val2
	case "ORI":
		result = val1 | val2
		fallthrough
	case "XORI":
		result = val1 ^ val2
	}
	e.RegisterFile[inst.Registers[0]] = result
}

func (e *Emulator) executeLoadUpperImmediate(inst *Instruction) {
	if inst.Registers[0] == 0 {
		return
	}

	val := uint32(inst.UnsignedConstant16) << 16
	e.RegisterFile[inst.Registers[0]] = val
}

func (e *Emulator) executeSetLessThan(inst *Instruction) {
	if inst.Registers[0] == 0 {
		return
	}

	val1 := e.RegisterFile[inst.Registers[1]]
	var val2 uint32
	if inst.Name == "SLTI" || inst.Name == "SLTIU" {
		val2 = uint32(inst.SignedConstant16)
	} else {
		val2 = e.RegisterFile[inst.Registers[2]]
	}

	var res bool
	switch inst.Name {
	case "SLTI", "SLT":
		res = int32(val1) < int32(val2)
	default:
		res = val1 < val2
	}

	if res {
		e.RegisterFile[inst.Registers[0]] = 1
	} else {
		e.RegisterFile[inst.Registers[0]] = 0
	}
}

func (e *Emulator) executeConstantShift(inst *Instruction) {
	if inst.Registers[0] == 0 {
		return
	}

	shiftAmount := inst.Constant5
	val := e.RegisterFile[inst.Registers[1]]

	var res uint32
	switch inst.Name {
	case "SLL":
		res = val << shiftAmount
	case "SRA":
		res = uint32(int32(val) >> shiftAmount)
	case "SRL":
		res = val >> shiftAmount
	}

	e.RegisterFile[inst.Registers[0]] = res
}

func (e *Emulator) executeRegisterShift(inst *Instruction) {
	if inst.Registers[0] == 0 {
		return
	}

	shiftAmount := e.RegisterFile[inst.Registers[2]]
	val := e.RegisterFile[inst.Registers[1]]

	var res uint32
	switch inst.Name {
	case "SLLV":
		res = val << shiftAmount
	case "SRAV":
		res = uint32(int32(val) >> shiftAmount)
	case "SRLV":
		res = val >> shiftAmount
	}

	e.RegisterFile[inst.Registers[0]] = res
}

func (e *Emulator) executeConditionalMove(inst *Instruction) {
	if inst.Registers[0] == 0 {
		return
	}

	condition := e.RegisterFile[inst.Registers[2]]

	var doMove bool
	switch inst.Name {
	case "MOVN":
		doMove = (condition != 0)
	case "MOVZ":
		doMove = (condition == 0)
	}

	if doMove {
		value := e.RegisterFile[inst.Registers[1]]
		e.RegisterFile[inst.Registers[0]] = value
	}
}
