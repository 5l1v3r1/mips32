package mips32

import "testing"

func TestEmulatorRegModifiers(t *testing.T) {
	code := `
		# Seed the program with two random numbers.
		LUI $1, 0xca6d
		ORI $1, $1, 0x8c46       # $r1 = 0xca6d8c46
		LUI $2, 0x0a93
		ORI $2, $2, 0xd70b       # $r2 = 0x0a93d70b

		ADDIU $3, $1, 0x1337
		ADDIU $3, $3, -0x10      # $r3 = 0xCA6D9F6D
		ADDU $4, $1, $2          # $r4 = 0xD5016351
		AND $5, $1, $2           # $r5 = 0xA018402
		ANDI $30, $1, 0xd70b     # $r30 = 0x8402
		MOVN $6, $1, $2          # $r6 = 0xca6d8c46
		MOVN $7, $1, $7          # $r7 = 0
		MOVZ $8, $1, $2          # $r8 = 0
		MOVZ $9, $1, $9          # $r9 = 0xca6d8c46
		NOR $10, $1, $2          # $r10 = 0x350020b0
		OR $11, $1, $2           # $r11 = 0xCAFFDF4F
		ORI $12, $1, 0xd70b      # $r12 = 0xCA6DDF4F
		SLL $13, $1, 11          # $r13 = 0x6c623000
		SLLV $14, $1, $2         # $r14 = 0x6c623000
		SLT $15, $1, $2          # $r15 = 1
		SLT $16, $2, $1          # $r16 = 0
		SLTU $17, $1, $2         # $r17 = 0
		SLTU $18, $2, $1         # $r18 = 1
		SLTI $19, $1, -1         # $r19 = 1
		SLTI $20, $2, -1         # $r20 = 0
		SLTIU $21, $1, -1        # $r21 = 1
		SLTIU $22, $2, -1        # $r22 = 1
		SRA $23, $1, 11          # $r23 = 0xFFF94DB1
		SRAV $24, $1, $2         # $r24 = 0xFFF94DB1
		SRL $25, $1, 11          # $r25 = 0x194DB1
		SRLV $26, $1, $2         # $r26 = 0x194DB1
		SUBU $27, $1, $2         # $r27 = 0xBFD9B53B
		XOR $28, $1, $2          # $r28 = 0xC0FE5B4D
		XORI $29, $1, 0xffff     # $r29 = 0xca6d73b9
	`
	lines, err := TokenizeSource(code)
	if err != nil {
		t.Fatal(err)
	}
	program, err := ParseExecutable(lines)
	if err != nil {
		t.Fatal(err)
	}
	memory := NewLazyMemory()
	emulator := &Emulator{
		Memory:     memory,
		Executable: program,
	}
	for !emulator.Done() {
		if err := emulator.Step(); err != nil {
			t.Fatal(err)
		}
	}
	regFile := RegisterFile{0, 0xca6d8c46, 0x0a93d70b, 0xCA6D9F6D, 0xD5016351, 0xA018402,
		0xca6d8c46, 0, 0, 0xca6d8c46, 0x350020b0, 0xCAFFDF4F, 0xCA6DDF4F, 0x6c623000, 0x6c623000, 1,
		0, 0, 1, 1, 0, 1, 1, 0xFFF94DB1, 0xFFF94DB1, 0x194DB1, 0x194DB1, 0xBFD9B53B, 0xC0FE5B4D,
		0xca6d73b9, 0x8402, 0}
	for i := 0; i < 32; i++ {
		if regFile[i] != emulator.RegisterFile[i] {
			t.Error("bad register", i, "-", emulator.RegisterFile[i])
		}
	}
}
