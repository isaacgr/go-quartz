package quartz

import (
	"fmt"
	"regexp"
)

var dotSRegexp *regexp.Regexp = regexp.MustCompile("[A-Z]{1,16}[0-9]{0,},[0-9]{0,}")

type Decoder[T any] func(line []byte) (*T, error)

type DecodeError struct {
	Cmd  string
	Line string
}

func (e *DecodeError) Error() string {
	return fmt.Sprintf(
		"Invalid arguments. Cmd [%s], Line [%s]",
		e.Cmd,
		e.Line,
	)
}

type DotSCmd struct {
	levels []string
	dst    string
	src    string
}

func DecodeDotS[T DotSCmd](line []byte) (*DotSCmd, error) {
	// cmd = VAB...1,1 or VAB...001,001
	ok := dotSRegexp.Match(line)
	if !ok {
		return nil, &DecodeError{".S", string(line)}
	}
	return &DotSCmd{}, nil
}
