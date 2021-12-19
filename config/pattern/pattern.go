package pattern

import (
	"errors"
	"fmt"
	"regexp"
)

var (
	parserRegex            = regexp.MustCompile("%[%a-zA-Z]")
	ErrInvalidPatternVar   = errors.New("invalid pattern variable")
	ErrMissingPatternValue = errors.New("missing value for pattern variable")
)

type Var rune

type Pattern string

type ParseResult struct {
	output string
	vars   map[Var]struct{}
}

func (r *ParseResult) String() string {
	return r.output
}

func (r *ParseResult) HasVar(variable Var) bool {
	_, ok := r.vars[variable]

	return ok
}

type Parser struct {
	variables map[Var]*string
}

func NewParser(variables map[Var]*string) Parser {
	return Parser{variables}
}

func (p Parser) Parse(pattern Pattern) (*ParseResult, error) {
	output := string(pattern)
	vars := make(map[Var]struct{})
	offset := 0

	for {
		matchRange := parserRegex.FindStringIndex(output[offset:])
		if matchRange == nil {
			break
		}

		start, end := matchRange[0]+offset, matchRange[1]+offset
		match := output[start:end]

		if match == "%%" {
			output = output[:start] + output[end-1:]
			offset = end - 1
			continue
		}

		variable := Var(match[1])
		value, ok := p.variables[variable]
		if !ok {
			return nil, fmt.Errorf("%w: %s", ErrInvalidPatternVar, string(variable))
		} else if value == nil {
			return nil, fmt.Errorf("%w: %s", ErrMissingPatternValue, string(variable))
		}

		vars[variable] = struct{}{}

		output = output[:start] + *value + output[end:]

		// We're adding the value but removing the variable
		offset = end + len(*value) - 2
	}

	return &ParseResult{output, vars}, nil
}

// ParseFirst parses the first pattern where none of the values are missing.
func (p Parser) ParseFirst(patterns []Pattern) (*ParseResult, error) {
	for _, pattern := range patterns {
		switch result, err := p.Parse(pattern); {
		case errors.Is(err, ErrMissingPatternValue):
			continue
		case err != nil:
			return nil, err
		default:
			return result, nil
		}
	}

	return nil, ErrMissingPatternValue
}
