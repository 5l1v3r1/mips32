package mips32

import "testing"

func TestInstCodingProgram(t *testing.T) {
	// TODO: add the rest of the instructions to this program.
	code := `
        NOP
        ADDIU $r5, $r4, -0x1337
        ADDU $r6, $r31, $r15
        AND $r31, $r5, $r1
        ANDI $r17, $r2, 0xf0f0

        LUI $r5, 0xf0f0
        MOVN $r17, $r18, $r19
        MOVZ $r18, $r19, $r20
        NOR $r1, $r2, $r3
        OR $r17, $r2, $r17

        ORI $r18, $r0, 0xf007
        SLL $r1, $r18, 7
        SLLV $r2, $r30, $r5
        SLT $r15, $r5, $r8
        SLTI $r15, $r5, -10

        SLTU $r0, $r5, $r8
        SLTIU $r15, $r5, -10
        SRA $r5, $r3, 15
        SRAV $r5, $r1, $r0
        SRL $r5, $r3, 15

        SRLV $r5, $r1, $r31
        SUBU $r9, $r10, $r31
        XOR $r2, $r3, $r4
        XORI $r2, $r3, 666
    `
	words := []uint32{
		0x00000000, 0x2485ECC9, 0x03ef3021, 0x00a1f824, 0x3051f0f0,
		0x3c05f0f0, 0x0253880b, 0x0274900a, 0x00430827, 0x00518825,
		0x3412f007, 0x001209c0, 0x00be1004, 0x00a8782a, 0x28affff6,
		0x00a8002b, 0x2caffff6, 0x00032bc3, 0x00012807, 0x00032bc2,
		0x03e12806, 0x015f4823, 0x00641026, 0x3862029a,
	}
	tokenizedLines, err := TokenizeSource(code)
	if err != nil {
		t.Fatal(err)
	}
	for i, line := range tokenizedLines {
		if line.Instruction == nil {
			t.Error("expecting instruction to be set for line", i)
			continue
		}
		inst, err := ParseTokenizedInstruction(line.Instruction)
		if err != nil {
			t.Error("failed to parse tokenized instruction for line", i)
			continue
		}
		encoded, err := inst.Encode(0, nil)
		if err != nil {
			t.Error("failed to encode instruction for line", i)
			continue
		}
		if encoded != words[i] {
			t.Error("incorrect encoding for instruction", i, "-", encoded)
		}
		decoded := DecodeInstruction(words[i])
		if !instructionsEquivalent(decoded, inst) {
			t.Error("incorrect decoding for instruction", i, "-", decoded)
		}
	}
}

func TestInstCodingControl(t *testing.T) {
	controls := map[string]bool{
		"BEQ $r15, $r17, SYM1": true,
		"BEQ $r15, $r17, SYM2": true,
		"J SYM3":               true,
		"J 0x10000004":         true,
		"BEQ $r15, $r17, SYM4": false,
		"BEQ $r13, $r31, SYM5": false,
		"J SYM1":               false,
		"BLEZ $15, 7":          false,
		"J 0x10000007":         false,
	}
	symbols := map[string]uint32{"SYM1": 0xffe0004, "SYM2": 0x10020000, "SYM3": 0x10050000,
		"SYM4": 0xffe0000, "SYM5": 0x10020004}
	for ctrl, success := range controls {
		tokens, err := TokenizeSource(ctrl)
		if err != nil {
			t.Error("control", ctrl, "-", err)
			continue
		}
		if len(tokens) != 1 || tokens[0].Instruction == nil {
			t.Error("bad tokens for", ctrl)
			continue
		}
		inst, err := ParseTokenizedInstruction(tokens[0].Instruction)
		if err != nil {
			t.Error("failed to parse tokenized instruction for", ctrl)
			continue
		}
		if _, err := inst.Encode(0x10000000, symbols); (err == nil) != success {
			t.Error("control", ctrl, "-", err)
		}
	}
}

func TestInstCodingRawWords(t *testing.T) {
	inst := DecodeInstruction(0xf2345678)
	if inst.Name != ".word" || inst.RawWord != 0xf2345678 {
		t.Error("did not properly decode word.")
	}

	i := &Instruction{Name: ".word", RawWord: 0xf2345678}
	if val, err := i.Encode(0, nil); err != nil {
		t.Error(err)
	} else if val != 0xf2345678 {
		t.Error("did not properly encode word.")
	}
}

func instructionsEquivalent(i1 *Instruction, i2 *Instruction) bool {
	if len(i1.Registers) != len(i2.Registers) {
		return false
	}
	for i, r := range i1.Registers {
		if i2.Registers[i] != r {
			return false
		}
	}
	if uint16(i1.SignedConstant16)|uint16(i1.UnsignedConstant16) !=
		uint16(i2.SignedConstant16)|uint16(i2.UnsignedConstant16) {
		return false
	}
	if i1.Name != i2.Name {
		return false
	}
	if i1.RawWord != i2.RawWord {
		return false
	}
	if i1.CodePointer != i2.CodePointer {
		return false
	}
	if i1.MemoryReference != i2.MemoryReference {
		return false
	}
	if i1.Constant5 != i2.Constant5 {
		return false
	}
	return true
}
