package mips32

import (
	"errors"
	"sort"
	"strconv"
)

// An Executable stores chunks of instructions (called segments) and a symbol table.
type Executable struct {
	// Segments maps chunks of instructions to various parts of the address space.
	Segments map[uint32][]Instruction

	// Symbols maps symbol names to their addresses.
	Symbols map[string]uint32
}

// ParseExecutable turns a tokenized source file into an executable blob.
//
// If the executable cannot be parsed for any reason, this will fail.
// Overlapping .text sections, invalid instructions, and repeated symbols will all cause errors.
func ParseExecutable(lines []TokenizedLine) (*Executable, error) {
	var segmentStart uint32
	var instructionAddr uint32
	res := &Executable{
		Segments: map[uint32][]Instruction{},
		Symbols:  map[string]uint32{},
	}
	for _, line := range lines {
		if line.Instruction != nil {
			parsed, err := ParseTokenizedInstruction(line.Instruction)
			if err != nil {
				return nil, errors.New("line " + strconv.Itoa(line.LineNumber) + ": " + err.Error())
			}
			if res.addressInUse(instructionAddr) {
				return nil, addressInUseError(line.LineNumber, instructionAddr)
			}
			res.Segments[segmentStart] = append(res.Segments[segmentStart], *parsed)
			instructionAddr += 4
		} else if line.Directive != nil {
			dir := line.Directive
			if dir.Name == "word" {
				if res.addressInUse(instructionAddr) {
					return nil, addressInUseError(line.LineNumber, instructionAddr)
				}
				nextInst := DecodeInstruction(dir.Constant)
				res.Segments[segmentStart] = append(res.Segments[segmentStart], *nextInst)
				instructionAddr += 4
			} else if dir.Name == "text" {
				if dir.Constant&3 != 0 {
					return nil, errors.New("line " + strconv.Itoa(line.LineNumber) +
						": misaligned segment")
				}
				segmentStart = dir.Constant
				instructionAddr = dir.Constant
			} else {
				return nil, errors.New("line " + strconv.Itoa(line.LineNumber) +
					": unknown directive: " + dir.Name)
			}
		} else if line.SymbolMarker != nil {
			sym := *line.SymbolMarker
			if _, ok := res.Symbols[sym]; ok {
				return nil, errors.New("line " + strconv.Itoa(line.LineNumber) +
					": repeated symbol declaration: " + sym)
			}
			res.Symbols[sym] = instructionAddr
		}
	}
	res.joinContiguousSegments()
	return res, nil
}

// addressInUse reports of a word-aligned address is being used by one of the segments.
func (e *Executable) addressInUse(addr uint32) bool {
	for segment, insts := range e.Segments {
		if segment <= addr && segment+uint32(len(insts)*4) > addr {
			return true
		}
	}
	return false
}

// joinContiguousSegments joins contiguous segments.
func (e *Executable) joinContiguousSegments() {
	l := make(uint32List, 0, len(e.Segments))
	for seg := range e.Segments {
		l = append(l, seg)
	}
	sort.Sort(l)

	for i := 0; i < len(l)-1; i++ {
		segStart := l[i]
		size := uint32(len(e.Segments[segStart]) * 4)
		if segStart+size == l[i+1] {
			nextInstructions := e.Segments[l[i+1]]
			e.Segments[segStart] = append(e.Segments[segStart], nextInstructions...)
			delete(e.Segments, l[i+1])
			copy(l[i:], l[i+1:])
			l = l[:len(l)-1]
			i--
		}
	}
}

func addressInUseError(line int, addr uint32) error {
	hexStr := "0x" + strconv.FormatUint(uint64(addr), 16)
	return errors.New("line " + strconv.Itoa(line) + ": overwriting address " + hexStr)
}

type uint32List []uint32

func (u uint32List) Len() int {
	return len(u)
}

func (u uint32List) Less(i, j int) bool {
	return u[i] < u[j]
}

func (u uint32List) Swap(i, j int) {
	u[i], u[j] = u[j], u[i]
}
