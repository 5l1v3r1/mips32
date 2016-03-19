package mips32

import "testing"

func TestParseTokenizedInstruction(t *testing.T) {
	cases := map[string]*Instruction{
		"ADDIU $r5, $r6, -17": &Instruction{
			Name:             "ADDIU",
			Registers:        []int{5, 6},
			SignedConstant16: -17,
		},
		"J FOO": &Instruction{
			Name: "J",
			CodePointer: CodePointer{
				Absolute: true,
				IsSymbol: true,
				Symbol:   "FOO",
			},
		},
	}
	for code, expected := range cases {
		tokenized, err := TokenizeSource(code)
		if err != nil {
			t.Error(code, "-", err)
			continue
		} else if len(tokenized) != 1 || tokenized[0].Instruction == nil {
			t.Error(code, "- bad tokens")
			continue
		}
		inst, err := ParseTokenizedInstruction(tokenized[0].Instruction)
		if err != nil {
			t.Error(code, "-", err)
			continue
		}
		if !instructionsEquivalent(inst, expected) {
			t.Error(code, "-", inst)
		}
	}

	broken := []string{"ADDIU $r5, $r6, $7", "J $r7"}
	for _, brokenCode := range broken {
		tokenized, err := TokenizeSource(brokenCode)
		if err != nil {
			t.Error(brokenCode, "-", err)
			continue
		} else if len(tokenized) != 1 || tokenized[0].Instruction == nil {
			t.Error(brokenCode, "- bad tokens")
			continue
		}
		if _, err := ParseTokenizedInstruction(tokenized[0].Instruction); err == nil {
			t.Error("expected error for", brokenCode)
		}
	}
}
