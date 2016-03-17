package mips32

type ArgumentType int

const (
	Register ArgumentType = iota
	Constant16
	Constant5
	CodePointer
	MemoryAddress
)

// A Template describes the kinds of arguments an instruction can take.
type Template struct {
	Name      string
	Arguments []ArgumentType
}

var Templates = []Template{
	{"NOP", []ArgumentType{}},
	{"ADDIU", []ArgumentType{Register, Register, Constant16}},
	{"ADDU", []ArgumentType{Register, Register, Register}},
	{"AND", []ArgumentType{Register, Register, Register}},
	{"ANDI", []ArgumentType{Register, Register, Constant16}},
	{"BEQ", []ArgumentType{Register, Register, CodePointer}},
	{"BGEZ", []ArgumentType{Register, CodePointer}},
	{"BGTZ", []ArgumentType{Register, CodePointer}},
	{"BLEZ", []ArgumentType{Register, CodePointer}},
	{"BLTZ", []ArgumentType{Register, CodePointer}},
	{"BNE", []ArgumentType{Register, Register, CodePointer}},
	{"J", []ArgumentType{CodePointer}},
	{"JAL", []ArgumentType{CodePointer}},
	{"JALR", []ArgumentType{Register}},
	{"JALR", []ArgumentType{Register, Register}},
	{"JR", []ArgumentType{Register}},
	{"LB", []ArgumentType{Register, MemoryAddress}},
	{"LBU", []ArgumentType{Register, MemoryAddress}},
	{"LW", []ArgumentType{Register, MemoryAddress}},
	{"SB", []ArgumentType{Register, MemoryAddress}},
	{"SW", []ArgumentType{Register, MemoryAddress}},
	{"LUI", []ArgumentType{Register, Constant16}},
	{"MOVN", []ArgumentType{Register, Register, Register}},
	{"MOVZ", []ArgumentType{Register, Register, Register}},
	{"OR", []ArgumentType{Register, Register, Register}},
	{"ORI", []ArgumentType{Register, Register, Constant16}},
	{"SLL", []ArgumentType{Register, Register, Constant5}},
	{"SLLV", []ArgumentType{Register, Register, Register}},
	{"SLT", []ArgumentType{Register, Register, Register}},
	{"SLTI", []ArgumentType{Register, Register, Constant5}},
	{"SLTIU", []ArgumentType{Register, Register, Constant5}},
	{"SLTU", []ArgumentType{Register, Register, Register}},
	{"SRA", []ArgumentType{Register, Register, Constant5}},
	{"SRAV", []ArgumentType{Register, Register, Register}},
	{"SRL", []ArgumentType{Register, Register, Constant5}},
	{"SRLV", []ArgumentType{Register, Register, Register}},
	{"SUBU", []ArgumentType{Register, Register, Register}},
	{"XOR", []ArgumentType{Register, Register, Register}},
	{"XORI", []ArgumentType{Register, Register, Constant5}},
}
