package pattern

import (
	"errors"
	"fmt"
	"regexp"
)

var (
	parserRegex          = regexp.MustCompile("%[%a-zA-Z]")
	ErrInvalidPatternVar = errors.New("invalid pattern variable")
)

type Var rune

type Pattern string

type Parser struct {
	variables map[Var]string
}

func NewParser(variables map[Var]string) Parser {
	return Parser{variables}
}

func (p Parser) Parse(pattern Pattern) (string, error) {
	result := string(pattern)
	offset := 0

	for {
		matchRange := parserRegex.FindStringIndex(result[offset:])
		if matchRange == nil {
			break
		}

		start, end := matchRange[0]+offset, matchRange[1]+offset
		match := result[start:end]

		if match == "%%" {
			result = result[:start] + result[end-1:]
			offset = end - 1
			continue
		}

		variable := Var(match[1])
		value, ok := p.variables[variable]
		if !ok {
			return "", fmt.Errorf("%w: %s", ErrInvalidPatternVar, string(variable))
		}

		result = result[:start] + value + result[end:]
		offset = end + len(value) - 2
	}

	return result, nil
}

func (p Parser) ParseMultiple(patterns []Pattern) ([]string, error) {
	var err error

	result := make([]string, len(patterns))
	for i, pattern := range patterns {
		result[i], err = p.Parse(pattern)
		if err != nil {
			return nil, err
		}
	}

	return result, nil
}
