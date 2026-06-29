package goquartz

import (
	"bytes"
	"fmt"
	"regexp"
	"strconv"
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
	dst    int
	src    int
}
type DstLockCmd struct {
	dst    int
	locked bool
}
type DstLockStatus struct {
	dst    int
	status bool
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
		dst, err := strconv.Atoi(matches[2])
		if err != nil {
			return nil, &DecodeError{".S", "Cannot infer destination index"}
		}
		src, err := strconv.Atoi(matches[3])
		if err != nil {
			return nil, &DecodeError{".S", "Cannot infer source index"}
		}
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
	resp := []*SetXptCmd{}
	// TODO: The plus sign case is confusing, because it seems to indicate
	// that a src can have a +, but I'm not sure how that makes sense since
	// its not appending any different type (like a new level)
	if bytes.Index(line, []byte("+")) > 0 {
		// Valid .M regex pattern, but we need to specifically test and extract
		// the case where + is used
		ok = dotMPlusRegexp.Match(line)
		if !ok {
			return nil, &DecodeError{".M", string(line)}
		}
		matches := dotMPlusRegexp.FindAllStringSubmatch(string(line), -1)
		if matches != nil {
			return nil, &DecodeError{".S", "Invalid request"}
		}

	}
	// It did compile using the other syntax
	matches := dotMRegexp.FindAllStringSubmatch(string(line), -1)
	if matches != nil {
		if len(matches) > 1 {
			// Must be the 'simple case' (i.e. .MV1,1,A1,1...)
			for _, m := range matches {
				dst, err := strconv.Atoi(m[2])
				if err != nil {
					return nil, &DecodeError{
						".M", "Cannot infer destination index",
					}
				}
				src, err := strconv.Atoi(m[3])
				if err != nil {
					return nil, &DecodeError{".M", "Cannot infer source index"}
				}
				resp = append(resp, &SetXptCmd{
					levels: []string{m[1]},
					src:    dst,
					dst:    src,
				})
			}
			return resp, nil
		}
		// Range case or like .S (i.e. .MV1,1)
		ok := dotSRegexp.Match(line)
		if !ok {
			// Might be the range case
			match := matches[0]
			levels := strings.Split(match[1], "")
			dstRange := strings.Split(match[2], "-")
			srcRange := strings.Split(match[3], "-")
			if len(dstRange) > 1 && len(srcRange) > 1 {
				if len(dstRange) != len(srcRange) {
					return nil, &DecodeError{
						".M", "Unequal source and dest ranges",
					}
				}
				for i, d := range dstRange {
					s := srcRange[i]
					dst, err := strconv.Atoi(d)
					if err != nil {
						return nil, &DecodeError{
							".M", "Cannot infer destination index",
						}
					}
					src, err := strconv.Atoi(s)
					if err != nil {
						return nil, &DecodeError{
							".M", "Cannot infer source index",
						}
					}
					resp = append(resp, &SetXptCmd{
						levels: levels,
						dst:    dst,
						src:    src,
					})
				}
				return resp, nil
			} else if len(dstRange) > 1 && len(srcRange) == 1 {
				for _, d := range dstRange {
					dst, err := strconv.Atoi(d)
					if err != nil {
						return nil, &DecodeError{
							".M", "Cannot infer destination index",
						}
					}
					src, err := strconv.Atoi(srcRange[0])
					if err != nil {
						return nil, &DecodeError{
							".M", "Cannot infer source index",
						}
					}
					resp = append(resp, &SetXptCmd{
						levels: levels,
						dst:    dst,
						src:    src,
					})
				}
				return resp, nil

			} else if len(srcRange) > 1 && len(dstRange) == 1 {
				for _, d := range dstRange {
					dst, err := strconv.Atoi(d)
					if err != nil {
						return nil, &DecodeError{
							".M", "Cannot infer destination index",
						}
					}
					src, err := strconv.Atoi(srcRange[0])
					if err != nil {
						return nil, &DecodeError{
							".M", "Cannot infer source index",
						}
					}
					resp = append(resp, &SetXptCmd{
						levels: levels,
						dst:    dst,
						src:    src,
					})
				}
				return resp, nil
			}
		}
		matches := dotSRegexp.FindStringSubmatch(string(line))
		if matches != nil {
			levels := strings.Split(matches[1], "")
			dst, err := strconv.Atoi(matches[2])
			if err != nil {
				return nil, &DecodeError{".M", "Cannot infer destination index"}
			}
			src, err := strconv.Atoi(matches[3])
			if err != nil {
				return nil, &DecodeError{".M", "Cannot infer source index"}
			}
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
