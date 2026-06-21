package goquartz

import (
	"fmt"
	"regexp"
)

var dotERegexp *regexp.Regexp = regexp.MustCompile("^[E]$")
var dotSRegexp *regexp.Regexp = regexp.MustCompile("^[A-Z]{1,16}[0-9]{0,},[0-9]{0,}$")

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

type DotEResp struct{}

type DotSCmd struct {
	levels []string
	dst    string
	src    string
}

func DecodeDotE(line []byte) (*DotEResp, error) {
	// .E
	ok := dotERegexp.Match(line)
	if !ok {
		return nil, &DecodeError{".E", string(line)}
	}
	return &DotEResp{}, nil
}

func DecodeDotS(line []byte) (*DotSCmd, error) {
	// cmd = VAB...1,1 or VAB...001,001
	ok := dotSRegexp.Match(line)
	if !ok {
		return nil, &DecodeError{".S", string(line)}
	}
	return &DotSCmd{}, nil
}
