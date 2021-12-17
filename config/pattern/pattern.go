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

type Parser struct {
	variables map[Var]*string
}

func NewParser(variables map[Var]*string) Parser {
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
		} else if value == nil {
			return "", fmt.Errorf("%w: %s", ErrMissingPatternValue, string(variable))
		}

		result = result[:start] + *value + result[end:]

		// We're adding the value but removing the variable
		offset = end + len(*value) - 2
	}

	return result, nil
}

// ParseFirst parses the first pattern where none of the values are missing.
func (p Parser) ParseFirst(patterns []Pattern) (string, error) {
	for _, pattern := range patterns {
		switch result, err := p.Parse(pattern); {
		case errors.Is(err, ErrMissingPatternValue):
			continue
		case err != nil:
			return "", err
		default:
			return result, nil
		}
	}

	return "", ErrMissingPatternValue
}
