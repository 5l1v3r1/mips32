package mips32

import "errors"

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
