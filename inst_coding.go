package mips32

var twoOperandImmediateOpcodes = map[uint32]string{
	0x09: "ADDIU",
	0x0c: "ANDI",
	0x0d: "ORI",
	0x0a: "SLTI",
	0x0b: "SLTIU",
	0x0e: "XORI",
}

var branchOpcodes = map[uint32]string{
	0x04: "BEQ",
	0x01: "BLTZ",
	0x07: "BGTZ",
	0x06: "BLEZ",
	0x05: "BNE",
}

var jTypeOpcodes = map[uint32]string{
	0x02: "J",
	0x03: "JAL",
}

var memoryOpcodes = map[uint32]string{
	0x20: "LB",
	0x24: "LBU",
	0x23: "LW",
	0x28: "SB",
	0x2b: "SW",
}

var constantShiftFuncs = map[uint32]string{
	0x00: "SLL",
	0x03: "SRA",
	0x02: "SRL",
}

var variableShiftFuncs = map[uint32]string{
	0x04: "SLLV",
	0x07: "SRAV",
	0x06: "SRLV",
}

var threeRegOperandFuncs = map[uint32]string{
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

const luiOpcode = 0x0f
const jrFunc = 0x08
const jalrFunc = 0x09

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

	if instName, ok := twoOperandImmediateOpcodes[opcode]; ok {
		return &Instruction{
			Name:               instName,
			Registers:          []int{registerT, registerS},
			UnsignedConstant16: uint16(immediate),
			SignedConstant16:   int16(immediate),
		}
	}

	if opcode == luiOpcode && registerS == 0 {
		return &Instruction{
			Name:               "LUI",
			Registers:          []int{registerT},
			UnsignedConstant16: uint16(immediate),
		}
	}

	if instName, ok := branchOpcodes[opcode]; ok {
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

	if instName, ok := jTypeOpcodes[opcode]; ok {
		jumpAddr := word & 0x03ffffff
		return &Instruction{
			Name:        instName,
			CodePointer: CodePointer{Absolute: true, Constant: jumpAddr},
		}
	}

	if instName, ok := memoryOpcodes[opcode]; ok {
		return &Instruction{
			Name:            instName,
			Registers:       []int{registerT},
			MemoryReference: MemoryReference{Register: registerS, Offset: int16(immediate)},
		}
	}

	if opcode == 0 {
		if instName, ok := constantShiftFuncs[funcField]; ok && registerS == 0 {
			return &Instruction{
				Name:      instName,
				Registers: []int{registerD, registerT},
				Constant5: shiftAmount,
			}
		}

		if instName, ok := variableShiftFuncs[funcField]; ok && shiftAmount == 0 {
			return &Instruction{
				Name:      instName,
				Registers: []int{registerD, registerT, registerS},
			}
		}

		if instName, ok := threeRegOperandFuncs[funcField]; ok && shiftAmount == 0 {
			return &Instruction{
				Name:      instName,
				Registers: []int{registerD, registerS, registerT},
			}
		}

		if opcode == 0 && registerT == 0 && registerD == 0 &&
			shiftAmount == 0 && funcField == jrFunc {
			return &Instruction{
				Name:      "JR",
				Registers: []int{registerS},
			}
		}

		if opcode == 0 && registerT == 0 && shiftAmount == 0 && funcField == jalrFunc {
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
