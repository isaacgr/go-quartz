package goquartz

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
)

var dotARegexp *regexp.Regexp = regexp.MustCompile("^[A]$")
var dotERegexp *regexp.Regexp = regexp.MustCompile("^[E]$")
var dotSRegexp *regexp.Regexp = regexp.MustCompile("^([A-Z]{1,17})([0-9]{1,}),([0-9]{1,})$")
var dotMRegexp *regexp.Regexp = regexp.MustCompile("([A-Za-z]+)?([A-Za-z0-9]+(?:[+-][A-Za-z0-9]+)*),([A-Za-z0-9]+(?:[+-][A-Za-z0-9]+)*)")
var dotMPlusRegexp *regexp.Regexp = regexp.MustCompile(`([A-Za-z]+)?(\d+)`)

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

type DotAResp struct{}
type DotEResp struct{}
type SetXptCmd struct {
	levels []string
	dst    string
	src    string
}

func DecodeDotA(line []byte) (*DotAResp, error) {
	// .A
	ok := dotARegexp.Match(line)
	if !ok {
		return nil, &DecodeError{".A", string(line)}
	}
	return &DotAResp{}, nil
}

func DecodeDotE(line []byte) (*DotEResp, error) {
	// .E
	ok := dotERegexp.Match(line)
	if !ok {
		return nil, &DecodeError{".E", string(line)}
	}
	return &DotEResp{}, nil
}

func DecodeDotS(line []byte) (*SetXptCmd, error) {
	// cmd = VAB...1,1 or VAB...001,001
	ok := dotSRegexp.Match(line)
	if !ok {
		return nil, &DecodeError{".S", string(line)}
	}
	matches := dotSRegexp.FindStringSubmatch(string(line))
	if matches != nil {
		levels := strings.Split(matches[1], "")
		dst := matches[2]
		src := matches[2]
		return &SetXptCmd{
			levels: levels,
			dst:    dst,
			src:    src,
		}, nil
	}
	return nil, &DecodeError{".S", "Invalid request"}
}

func DecodeDotM(line []byte) ([]*SetXptCmd, error) {
	ok := dotMRegexp.Match(line)
	if !ok {
		return nil, &DecodeError{".M", string(line)}
	}
	if bytes.Index(line, []byte("+")) > 0 {
		// Valid .M regex pattern, but we need to specifically test and extract
		// the case where + is used
		ok = dotMPlusRegexp.Match(line)
		if !ok {
			return nil, &DecodeError{".M", string(line)}
		}
		matches := dotMPlusRegexp.FindAllStringSubmatch(string(line), -1)
		if matches != nil {

		}

	}
	// It did compile using the other syntax
	matches := dotMRegexp.FindAllStringSubmatch(string(line), -1)
	if matches != nil {
		resp := []*SetXptCmd{}
		if len(matches) > 1 {
			// Must be the 'simple case' (i.e. .MV1,1,A1,1...)
			for _, m := range matches {
				resp = append(resp, &SetXptCmd{
					levels: []string{m[1]},
					src:    m[2],
					dst:    m[3],
				})
			}
			return resp, nil
		}
		// Range case or like .S (i.e. .MV1,1)
		ok := dotSRegexp.Match(line)
		if !ok {
			// Might be the range case
			return nil, &DecodeError{".M", "Invalid request"}
		}
		matches := dotSRegexp.FindStringSubmatch(string(line))
		if matches != nil {
			levels := strings.Split(matches[1], "")
			dst := matches[2]
			src := matches[2]
			resp = append(resp, &SetXptCmd{
				levels: levels,
				src:    src,
				dst:    dst,
			})
			return resp, nil
		}
	}

	return nil, &DecodeError{".M", "Invalid request"}
}
