package mips32

import (
	"errors"
	"regexp"
	"strconv"
	"strings"
)

var (
	registerNamePattern = "\\$(t[0-9]|s[0-7]|a[0-3]|v[01]|sp|fp|gp|ra|zero|" +
		"r?([0-9]|[0-2][0-9]|3[01]))"
	constantNumberPattern = "(-?[0-9]*|-?0x[0-9a-fA-F]*)"
	symbolNamePattern     = "([a-zA-Z0-9_]*)"
)

var (
	registerRegexp = regexp.MustCompile("^" + registerNamePattern + "$")
	constantRegexp = regexp.MustCompile("^" + constantNumberPattern + "$")
	symbolRegexp   = regexp.MustCompile("^" + symbolNamePattern + "$")
	memoryRegexp   = regexp.MustCompile("^(" + constantNumberPattern + "|)\\(" +
		registerNamePattern + "\\)$")
	memorySubfieldsRegexp = regexp.MustCompile("^(.*)\\((.*)\\)$")
)

// A CodePointer contains some kind of information indicating where a piece of code is.
type CodePointer struct {
	// Absolute is true if this is an AbsoluteCodePointer.
	Absolute bool

	// IsSymbol is true if this is represented by a symbol rather than a constant.
	IsSymbol bool

	Symbol   string
	Constant uint32
}

// A MemoryReference stores information about a memory reference.
type MemoryReference struct {
	Register int
	Offset   int16
}

// An ArgToken represents a register, a number, a symbol, or a memory location.
// For instance, the instruction "SB $5, 5($6)" contains three tokens.
//
// An ArgToken may be able to serve as multiple types of arguments.
// For example, the ArgToken for "0x5" could be a Constant5, a Constant16, or a CodePointer.
type ArgToken struct {
	isRegister bool
	register   int

	isConstant bool
	constant   uint32

	isSymbol bool
	symbol   string

	isMemory    bool
	memRegister int
	memOffset   int16
}

// ParseArgToken parses a human-readable token string.
// The string must not have any leading or trailing whitespace.
func ParseArgToken(tokenStr string) (token *ArgToken, err error) {
	if registerRegexp.MatchString(tokenStr) {
		return parseRegisterArgToken(tokenStr)
	} else if constantRegexp.MatchString(tokenStr) {
		return parseConstantArgToken(tokenStr)
	} else if symbolRegexp.MatchString(tokenStr) {
		return parseSymbolArgToken(tokenStr)
	} else if memoryRegexp.MatchString(tokenStr) {
		return parseMemoryArgToken(tokenStr)
	}
	return nil, errors.New("unable to parse token: " + tokenStr)
}

// Register returns the register index represented by this token.
// If this token cannot be treated as a register index, ok will be false.
func (t *ArgToken) Register() (regIndex int, ok bool) {
	return t.register, t.isRegister
}

// UnsignedConstant16 returns the 16-bit zero-extended constant represented by this token.
// If this token cannot be treated as an unsigned 16-bit constant, ok will be false.
func (t *ArgToken) UnsignedConstant16() (constant uint16, ok bool) {
	return uint16(t.constant), t.isConstant && (t.constant&0xffff0000) == 0
}

// SignedConstant16 returns the 16-bit sign-extended constant represented by this token.
// If this token cannot be treated as a signed 16-bit constant, ok will be false.
func (t *ArgToken) SignedConstant16() (constant int16, ok bool) {
	constant = int16(t.constant)
	ok = t.isConstant && t.constant == uint32(constant)
	return
}

// Constant5 returns the 5-bit unsigned constant represented by this token.
// If this token cannot be treated as a 5-bit constant, ok will be false.
func (t *ArgToken) Constant5() (constant uint8, ok bool) {
	return uint8(t.constant), t.isConstant && t.constant < 0x20
}

// RelativeCodePointer returns the relative code pointer represented by this token.
// If this token cannot be treated as a relative code pointer, ok will be false.
//
// If the code pointer is a constant, the constant value will represent an 18-bit signed address
// which is signed extended to 32 bits.
func (t *ArgToken) RelativeCodePointer() (ptr CodePointer, ok bool) {
	if t.isConstant {
		constant := int16(t.constant >> 2)
		if uint32(constant)<<2 != t.constant&0xfffffffc {
			return
		}
		return CodePointer{Constant: t.constant}, true
	} else if t.isSymbol {
		return CodePointer{IsSymbol: true, Symbol: t.symbol}, true
	} else {
		return
	}
}

// AbsoluteCodePointer returns the absolute code pointer represented by this token.
// If this token cannot be treated as an absolute code pointer, ok will be false.
//
// If the code pointer is a constant, it will be an absolute jump destination.
// The destination address should include the high bits of the intended PC+4 value.
func (t *ArgToken) AbsoluteCodePointer() (ptr CodePointer, ok bool) {
	if t.isConstant {
		return CodePointer{Absolute: true, Constant: uint32(t.constant)}, true
	} else if t.isSymbol {
		return CodePointer{Absolute: true, IsSymbol: true, Symbol: t.symbol}, true
	} else {
		return
	}
}

// MemoryReference returns the MemoryReference represented by this token.
// If this token cannot be treated as a MemoryReference, ok will be false.
func (t *ArgToken) MemoryReference() (ref MemoryReference, ok bool) {
	return MemoryReference{Register: t.memRegister, Offset: t.memOffset}, t.isMemory
}

func parseRegisterArgToken(tokenStr string) (token *ArgToken, err error) {
	regNum, err := parseRegister(tokenStr)
	if err != nil {
		return nil, err
	} else {
		return &ArgToken{isRegister: true, register: regNum}, nil
	}
}

func parseConstantArgToken(tokenStr string) (token *ArgToken, err error) {
	num, err := parseConstant(tokenStr)
	if err != nil {
		return nil, err
	}
	return &ArgToken{isConstant: true, constant: num}, nil
}

func parseSymbolArgToken(tokenStr string) (token *ArgToken, err error) {
	return &ArgToken{isSymbol: true, symbol: tokenStr}, nil
}

func parseMemoryArgToken(tokenStr string) (token *ArgToken, err error) {
	pieces := memorySubfieldsRegexp.FindStringSubmatch(tokenStr)
	if pieces == nil {
		return nil, errors.New("invalid memory reference: " + tokenStr)
	}

	reg, err := parseRegister(pieces[2])
	if err != nil {
		return
	}

	var offset int16
	if len(pieces[1]) != 0 {
		if offNum, err := parseConstant(pieces[1]); err != nil {
			return nil, err
		} else if (offNum&0xffff8000) != 0xffff8000 && (offNum&0xffff8000) != 0 {
			return nil, errors.New("memory offset out of bounds: " + pieces[1])
		} else {
			offset = int16(offNum)
		}
	}

	return &ArgToken{isMemory: true, memOffset: offset, memRegister: reg}, nil
}

func parseRegister(tokenStr string) (regIndex int, err error) {
	if !strings.HasPrefix(tokenStr, "$") {
		return 0, errors.New("missing $ in register name: " + tokenStr)
	}
	rawName := tokenStr[1:]
	specialCases := map[string]int{"zero": 0, "sp": 29, "fp": 30, "gp": 28, "ra": 31}
	if num, ok := specialCases[rawName]; ok {
		return num, nil
	}

	prefixes := map[string]int{"t([0-7])": 8, "t([89])": 24 - 8, "s([0-7])": 16, "a([0-3])": 4,
		"v([01])": 2, "r([12][0-9]|3[01]|[0-9])": 0}
	for prefix, start := range prefixes {
		r := regexp.MustCompile("^" + prefix + "$")
		m := r.FindStringSubmatch(rawName)
		if m != nil {
			idx, err := strconv.Atoi(m[1])
			if err != nil {
				panic("failed to parse integer component of register name")
			}
			return idx + start, nil
		}
	}

	rawNum, err := strconv.Atoi(rawName)
	if err != nil {
		return 0, err
	} else if rawNum < 0 || rawNum >= 0x20 {
		return 0, errors.New("invalid register index: " + rawName)
	}
	return rawNum, nil
}

func parseConstant(tokenStr string) (constant uint32, err error) {
	resNum, err := strconv.ParseInt(tokenStr, 0, 64)
	if err != nil {
		return 0, err
	}
	if resNum > 0xffffffff || resNum < -0xffffffff {
		return 0, err
	}
	return uint32(resNum), nil
}
