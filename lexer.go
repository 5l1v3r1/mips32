package mips32

import (
	"errors"
	"regexp"
	"strconv"
	"strings"
)

var (
	commentRegexp      = regexp.MustCompile("^(.*?)(#|//|;)(.*)$")
	directiveRegexp    = regexp.MustCompile("^\\.(text|data)\\s+" + constantNumberPattern + "$")
	symbolMarkerRegexp = regexp.MustCompile("^" + symbolNamePattern + ":$")
	instNameRegexp     = regexp.MustCompile("^[A-Z]*$")
)

// A TokenizedLine represents one line of an assembly program, translated into syntactic tokens.
// No more than one of Directive, SymbolDecl, and Instruction will be non-nil.
// The Comment field may be non-nil regardless of the other fields.
type TokenizedLine struct {
	LineNumber int
	Comment    *string

	Directive    *TokenizedDirective
	Instruction  *TokenizedInstruction
	SymbolMarker *string
}

// A TokenizedDirective represents a directive like ".text 0x5000" or ".data 0x0".
type TokenizedDirective struct {
	Name     string
	Constant uint32
}

// A TokenizedInstruction represents an instruction call.
type TokenizedInstruction struct {
	Name string
	Args []*ArgToken
}

// TokenizeSource takes a source file and tokenizes each line.
// It returns an array of tokenized lines, on an error if one occurred.
func TokenizeSource(source string) ([]TokenizedLine, error) {
	splitLines := strings.Split(source, "\n")
	res := make([]TokenizedLine, 0, len(splitLines))
	for lineNum, lineText := range splitLines {
		line, err := tokenizeLine(lineText)
		if err != nil {
			linePreamble := "error on line " + strconv.Itoa(lineNum+1) + ": "
			return nil, errors.New(linePreamble + err.Error())
		}
		line.LineNumber = lineNum + 1
		res = append(res, line)
	}
	return res, nil
}

// tokenizeLine tokenizes a single line of assembly code.
func tokenizeLine(lineText string) (line TokenizedLine, err error) {
	trimmed := strings.TrimSpace(lineText)
	if len(trimmed) == 0 {
		return
	}

	commentMatch := commentRegexp.FindStringSubmatch(lineText)
	if commentMatch != nil {
		commentStr := commentMatch[2]
		beforeComment := commentMatch[1]
		line, err = tokenizeLine(beforeComment)
		line.Comment = &commentStr
		return
	}

	directiveMatch := commentRegexp.FindStringSubmatch(lineText)
	if directiveMatch != nil {
		directiveConstant, err := parseConstant(directiveMatch[2])
		if err != nil {
			return line, err
		}
		return TokenizedLine{
			Directive: &TokenizedDirective{
				Name:     directiveMatch[1],
				Constant: directiveConstant,
			},
		}, nil
	}

	fields := strings.Fields(lineText)
	if len(fields) == 0 || !instNameRegexp.MatchString(fields[0]) {
		err = errors.New("unknown instruction")
		return
	}

	line.Instruction = &TokenizedInstruction{
		Name: fields[0],
		Args: make([]*ArgToken, len(fields)-1),
	}

	for i, field := range fields[1:] {
		if i != len(fields)-2 {
			if !strings.HasSuffix(field, ",") {
				err = errors.New("missing comma after operand " + strconv.Itoa(i+1))
				return
			}
		}
		line.Instruction.Args[i], err = ParseArgToken(field)
		if err != nil {
			err = errors.New("operand " + strconv.Itoa(i+1) + ": " + err.Error())
			return
		}
	}

	return
}
