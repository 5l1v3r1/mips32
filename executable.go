package mips32

import "errors"

type Executable struct {
	// Segments maps chunks of instructions to various parts of the address space.
	Segments map[uint32][]Instruction
}

// Instruction stores all of the information about an instruction in distinct fields.
// This makes it possible to execute an instruction and see exactly what its operands are.
type Instruction struct {
	// Name is the instruction's name.
	//
	// If this instruction is from a ".word" directive which does not correspond to a valid
	// instruction, then the Name field is ".word" and the RawWord field is set.
	Name string

	// Registers is the list of register indices passed to this instruction.
	// This list is in the same order as the instruction's operands in assembly.
	Registers []int

	UnsignedConstant16 uint16
	SignedConstant16   int16
	Constant5          uint8
	CodePointer        CodePointer
	MemoryReference    MemoryReference

	// RawWord is only used for instructions which cannot be decoded.
	// This is only used when Name is set to ".word"
	RawWord uint32
}

// DecodeInstruction returns an Instruction for a 32-bit word.
// This can never fail, since invalid instructions can be treated as ".word" directives.
func DecodeInstruction(word uint32) *Instruction {
	opcode := (word >> 26) & 0x3f
	registerS := int((word >> 21) & 0x1f)
	registerT := int((word >> 16) & 0x1f)
	registerD := int((word >> 11) & 0x1f)
	shiftAmount := uint8((word >> 6) & 0x1f)
	funcField := word & 0x3f
	immediate := word & 0xffff

	twoOperandImmediates := map[uint32]string{
		0x09: "ADDIU",
		0x0c: "ANDI",
		0x0d: "ORI",
		0x0a: "SLTI",
		0x0b: "SLTIU",
		0x0e: "XORI",
	}
	if instName, ok := twoOperandImmediates[opcode]; ok {
		return &Instruction{
			Name:               instName,
			Registers:          []int{registerT, registerS},
			UnsignedConstant16: uint16(immediate),
			SignedConstant16:   int16(immediate),
		}
	}

	if opcode == 0x0f && registerS == 0 {
		return &Instruction{
			Name:               "LUI",
			Registers:          []int{registerT},
			UnsignedConstant16: uint16(immediate),
		}
	}

	branches := map[uint32]string{
		0x04: "BEQ",
		0x01: "BLTZ",
		0x07: "BGTZ",
		0x06: "BLEZ",
		0x05: "BNE",
	}
	if instName, ok := branches[opcode]; ok {
		if instName == "BEQ" || instName == "BNE" {
			return &Instruction{
				Name:        instName,
				Registers:   []int{registerS, registerT},
				CodePointer: CodePointer{Constant: immediate},
			}
		}
		if instName == "BLTZ" && registerT == 1 {
			instName = "BGTZ"
			registerT = 0
		}
		if registerT == 0 {
			return &Instruction{
				Name:        instName,
				Registers:   []int{registerS},
				CodePointer: CodePointer{Constant: immediate},
			}
		}
	}

	jTypes := map[uint32]string{0x02: "J", 0x03: "JAL"}
	if instName, ok := jTypes[opcode]; ok {
		jumpAddr := word & 0x03ffffff
		return &Instruction{
			Name:        instName,
			CodePointer: CodePointer{Absolute: true, Constant: jumpAddr},
		}
	}

	memoryInstructions := map[uint32]string{
		0x20: "LB",
		0x24: "LBU",
		0x23: "LW",
		0x28: "SB",
		0x2b: "SW",
	}
	if instName, ok := memoryInstructions[opcode]; ok {
		return &Instruction{
			Name:            instName,
			Registers:       []int{registerT},
			MemoryReference: MemoryReference{Register: registerS, Offset: int16(immediate)},
		}
	}

	if opcode == 0 {
		constantShifts := map[uint32]string{
			0x00: "SLL",
			0x03: "SRA",
			0x02: "SRL",
		}
		if instName, ok := constantShifts[funcField]; ok && registerS == 0 {
			return &Instruction{
				Name:      instName,
				Registers: []int{registerD, registerT},
				Constant5: shiftAmount,
			}
		}

		variableShifts := map[uint32]string{
			0x04: "SLLV",
			0x07: "SRAV",
			0x06: "SRLV",
		}
		if instName, ok := variableShifts[funcField]; ok && shiftAmount == 0 {
			return &Instruction{
				Name:      instName,
				Registers: []int{registerD, registerT, registerS},
			}
		}

		threeRegOperands := map[uint32]string{
			0x21: "ADDU",
			0x24: "AND",
			0x0b: "MOVN",
			0x0a: "MOVZ",
			0x27: "NOR",
			0x25: "OR",
			0x2a: "SLT",
			0x2b: "SLTU",
			0x23: "SUBU",
			0x26: "XOR",
		}
		if instName, ok := threeRegOperands[funcField]; ok && shiftAmount == 0 {
			return &Instruction{
				Name:      instName,
				Registers: []int{registerD, registerS, registerT},
			}
		}

		if opcode == 0 && registerT == 0 && registerD == 0 &&
			shiftAmount == 0 && funcField == 0x08 {
			return &Instruction{
				Name:      "JR",
				Registers: []int{registerS},
			}
		}

		if opcode == 0 && registerT == 0 && shiftAmount == 0 && funcField == 0x09 {
			return &Instruction{
				Name:      "JALR",
				Registers: []int{registerD, registerS},
			}
		}
	}

	return &Instruction{
		Name:    ".word",
		RawWord: word,
	}
}

// ParseTokenizedInstruction generates an Instruction which represents a TokenizedInstruction.
// This may fail if the instruction is invalid, in which case an error is returned.
func ParseTokenizedInstruction(t *TokenizedInstruction) (*Instruction, error) {
	validName := false
	for _, template := range Templates {
		if template.Name == t.Name {
			validName = true
		}
		if template.Match(t) {
			res := &Instruction{Name: t.Name}
			for i, arg := range template.Arguments {
				tokArg := t.Arguments[i]
				switch arg {
				case Register:
					reg, _ := tokArg.Register()
					res.Registers = append(res.Registers, reg)
				case SignedConstant16:
					res.SignedConstant16, _ = tokArg.SignedConstant16()
				case UnsignedConstant16:
					res.UnsignedConstant16, _ = tokArg.UnsignedConstant16()
				case Constant5:
					res.Constant5, _ = tokArg.Constant5()
				case AbsoluteCodePointer:
					res.CodePointer, _ = tokArg.AbsoluteCodePointer()
				case RelativeCodePointer:
					res.CodePointer, _ = tokArg.RelativeCodePointer()
				case MemoryAddress:
					res.MemoryReference, _ = tokArg.MemoryReference()
				}
			}
			return res, nil
		}
	}
	if validName {
		return nil, errors.New("bad instruction usage for " + t.Name)
	} else {
		return nil, errors.New("unknown instruction: " + t.Name)
	}
}
