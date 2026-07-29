package goquartz

import (
	"bytes"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var dotARegexp *regexp.Regexp = regexp.MustCompile(`([A-Z]+)([0-9]+),([0-9]+)`)
var dotERegexp *regexp.Regexp = regexp.MustCompile("^[E]$")
var dotSRegexp *regexp.Regexp = regexp.MustCompile("^([A-Z]{1,17})([0-9]{1,}),([0-9]{1,})$")
var dotMRegexp *regexp.Regexp = regexp.MustCompile("([A-Za-z]+)?([A-Za-z0-9]+(?:[+-][A-Za-z0-9]+)*),([A-Za-z0-9]+(?:[+-][A-Za-z0-9]+)*)")
var dotMPlusRegexp *regexp.Regexp = regexp.MustCompile(`([A-Za-z]+)?(\d+)`)
var dotBRegexp *regexp.Regexp = regexp.MustCompile("^([LUIA])([0-9]{1,})(,[0-9]{1,})?$")
var dotFRegexp *regexp.Regexp = regexp.MustCompile("^[0-9]{1,3}$")
var dotIRegexp *regexp.Regexp = regexp.MustCompile("^([A-Z]{1})([0-9]{1,})$")
var dotLRegexp *regexp.Regexp = regexp.MustCompile("^([A-Z]{1})([0-9]{1,},-?)([0-9]{1,})?$")
var dotRRegexp *regexp.Regexp = regexp.MustCompile(`^([DSLETMA]{1,2})([0-9A-Z],)?(.*)$`)
var dotWRegexp *regexp.Regexp = regexp.MustCompile(`^([DSLETM]{1,2})([0-9A-Z]),(.*)$`)
var dotQRegexp *regexp.Regexp = regexp.MustCompile(`^([CRFSDL])([0-9]{1,})(?:(?:([A-Z]{1})([0-9]{1,}),([0-9]{1,}))|(T1\:[0-9]{2}\:[0-9]{2}\:[0-9]{2}\:[0-9]{2})|,([0-9]{1,}))?$`)

type DecodeError struct {
	Cmd  string
	Line string
	Msg  string
}

func (e *DecodeError) Error() string {
	return fmt.Sprintf(
		"Unable to decode command. Cmd [%s], Line [%s], Error [%s]",
		e.Cmd,
		e.Line,
		e.Msg,
	)
}

type DotAResp struct {
	Level string
	Dst   int
	Src   int
}
type DotEResp struct{}
type SetXptCmd struct {
	Levels []string
	Dst    int
	Src    int
}
type DstLockCmd struct {
	Dst    int
	Type   string
	Locked int
}
type DotFCmd struct {
	Salvo int
}
type DotICmd struct {
	Level string
	Dst   int
}
type DotLCmd struct {
	Level string
	Dst   int
	Src   int
}
type DotRCmd struct {
	Mnemonic string
	Type     string
	DstSrc   int
	Level    string
}
type DotWCmd struct {
	Mnemonic string
	Type     string
	DstSrc   int
	Level    string
}
type DotQFTime struct {
	Hours   int
	Minutes int
	Seconds int
	Frames  int
}
type DotQCmd struct {
	Type     string
	Salvo    int
	Level    string
	Dest     int
	Src      int
	Count    int
	Response bool
	FTime    *DotQFTime
}

func DecodeDotA(line []byte) ([]DotAResp, error) {
	// .AV1,1
	ok := dotARegexp.Match(line)
	if !ok {
		return nil, &DecodeError{
			Cmd:  ".A",
			Line: string(line),
			Msg:  "Invalid format",
		}
	}
	resp := []DotAResp{}
	matches := dotARegexp.FindAllStringSubmatch(string(line), -1)
	if matches != nil {
		for _, match := range matches {
			if len(match[1]) > 1 {
				return nil, &DecodeError{
					Cmd:  ".A",
					Line: string(line),
					Msg:  "Too many levels in response message",
				}

			}
			level := match[1]

			dst, err := strconv.Atoi(match[2])
			if err != nil {
				return nil, &DecodeError{
					Cmd:  ".A",
					Line: string(line),
					Msg:  "Cannot infer dst index",
				}
			}
			src, err := strconv.Atoi(match[3])
			if err != nil {
				return nil, &DecodeError{
					Cmd:  ".A",
					Line: string(line),
					Msg:  "Cannot infer src index",
				}
			}
			resp = append(resp, DotAResp{
				Level: level,
				Src:   src,
				Dst:   dst,
			})
		}
		return resp, nil
	}
	return nil, &DecodeError{
		Cmd:  ".A",
		Line: string(line),
		Msg:  "Invalid response format",
	}
}

func DecodeDotE(line []byte) (*DotEResp, error) {
	// .E
	ok := dotERegexp.Match(line)
	if !ok {
		return nil, &DecodeError{
			Cmd:  ".E",
			Line: string(line),
			Msg:  "Invalid format",
		}
	}
	return &DotEResp{}, nil
}

func DecodeDotS(line []byte) (*SetXptCmd, error) {
	// cmd = VAB...1,1 or VAB...001,001
	ok := dotSRegexp.Match(line)
	if !ok {
		return nil, &DecodeError{
			Cmd:  ".S",
			Line: string(line),
			Msg:  "Invalid format",
		}
	}
	matches := dotSRegexp.FindStringSubmatch(string(line))
	if matches != nil {
		levels := strings.Split(matches[1], "")
		dst, err := strconv.Atoi(matches[2])
		if err != nil {
			return nil, &DecodeError{
				Cmd:  ".S",
				Line: string(line),
				Msg:  "Cannot infer dst index",
			}
		}
		src, err := strconv.Atoi(matches[3])
		if err != nil {
			return nil, &DecodeError{
				Cmd:  ".S",
				Line: string(line),
				Msg:  "Cannot infer src index",
			}
		}
		return &SetXptCmd{
			Levels: levels,
			Dst:    dst,
			Src:    src,
		}, nil
	}
	return nil, &DecodeError{
		Cmd:  ".S",
		Line: string(line),
		Msg:  "Invalid request",
	}
}

func DecodeDotM(line []byte) ([]*SetXptCmd, error) {
	ok := dotMRegexp.Match(line)
	if !ok {
		return nil, &DecodeError{
			Cmd:  ".M",
			Line: string(line),
			Msg:  "Invalid format",
		}
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
			return nil, &DecodeError{
				Cmd:  ".M",
				Line: string(line),
				Msg:  "Invalid (+) syntax",
			}
		}
		matches := dotMPlusRegexp.FindAllStringSubmatch(string(line), -1)
		if matches != nil {
			// TODO
		}
	} else {
		// It did compile using the other syntax
		matches := dotMRegexp.FindAllStringSubmatch(string(line), -1)
		if matches != nil {
			if len(matches) > 1 {
				// Must be the 'simple case' (i.e. .MV1,1,A1,1...)
				for _, m := range matches {
					dst, err := strconv.Atoi(m[2])
					if err != nil {
						return nil, &DecodeError{
							Cmd:  ".M",
							Line: string(line),
							Msg:  "Cannot infer dst index",
						}
					}
					src, err := strconv.Atoi(m[3])
					if err != nil {
						return nil, &DecodeError{
							Cmd:  ".M",
							Line: string(line),
							Msg:  "Cannot infer src index",
						}
					}
					resp = append(resp, &SetXptCmd{
						Levels: []string{m[1]},
						Src:    dst,
						Dst:    src,
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
							Cmd:  ".M",
							Line: string(line),
							Msg:  "Unequal src/dst ranges",
						}
					}
					for i, d := range dstRange {
						s := srcRange[i]
						dst, err := strconv.Atoi(d)
						if err != nil {
							return nil, &DecodeError{
								Cmd:  ".M",
								Line: string(line),
								Msg:  "Cannot infer dst index",
							}
						}
						src, err := strconv.Atoi(s)
						if err != nil {
							return nil, &DecodeError{
								Cmd:  ".M",
								Line: string(line),
								Msg:  "Cannot infer src index",
							}
						}
						resp = append(resp, &SetXptCmd{
							Levels: levels,
							Dst:    dst,
							Src:    src,
						})
					}
					return resp, nil
				} else if len(dstRange) > 1 && len(srcRange) == 1 {
					for _, d := range dstRange {
						dst, err := strconv.Atoi(d)
						if err != nil {
							return nil, &DecodeError{
								Cmd:  ".M",
								Line: string(line),
								Msg:  "Cannot infer dst index",
							}
						}
						src, err := strconv.Atoi(srcRange[0])
						if err != nil {
							return nil, &DecodeError{
								Cmd:  ".M",
								Line: string(line),
								Msg:  "Cannot infer src index",
							}
						}
						resp = append(resp, &SetXptCmd{
							Levels: levels,
							Dst:    dst,
							Src:    src,
						})
					}
					return resp, nil

				} else if len(srcRange) > 1 && len(dstRange) == 1 {
					for _, d := range dstRange {
						dst, err := strconv.Atoi(d)
						if err != nil {
							return nil, &DecodeError{
								Cmd:  ".M",
								Line: string(line),
								Msg:  "Cannot infer dst index",
							}
						}
						src, err := strconv.Atoi(srcRange[0])
						if err != nil {
							return nil, &DecodeError{
								Cmd:  ".M",
								Line: string(line),
								Msg:  "Cannot infer src index",
							}
						}
						resp = append(resp, &SetXptCmd{
							Levels: levels,
							Dst:    dst,
							Src:    src,
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
					return nil, &DecodeError{
						Cmd:  ".M",
						Line: string(line),
						Msg:  "Cannot infer dst index",
					}
				}
				src, err := strconv.Atoi(matches[3])
				if err != nil {
					return nil, &DecodeError{
						Cmd:  ".M",
						Line: string(line),
						Msg:  "Cannot infer src index",
					}
				}
				resp = append(resp, &SetXptCmd{
					Levels: levels,
					Src:    src,
					Dst:    dst,
				})
				return resp, nil
			}
		}

	}
	return nil, &DecodeError{
		Cmd:  ".M",
		Line: string(line),
		Msg:  "Invalid request",
	}
}

func DecodeDotB(line []byte) (*DstLockCmd, error) {
	// .BL1 OR .BU1 OR .BI1 OR .BA1,<0-255>
	ok := dotBRegexp.Match(line)
	if !ok {
		return nil, &DecodeError{
			Cmd:  ".B",
			Line: string(line),
			Msg:  "Invalid format",
		}
	}
	matches := dotBRegexp.FindStringSubmatch(string(line))
	if matches != nil {
		var dest int
		var err error
		cmdType := matches[1]
		dest, err = strconv.Atoi(matches[2])
		if err != nil {
			return nil, &DecodeError{
				Cmd:  ".B",
				Line: string(line),
				Msg:  "Cannot parse dst",
			}
		}

		cmd := &DstLockCmd{
			Dst:  dest,
			Type: cmdType,
		}

		if cmdType == "A" {
			if len(matches) < 4 {
				return nil, &DecodeError{
					Cmd:  ".B",
					Line: string(line),
					Msg:  "Missing lock state",
				}
			}
			locked, err := strconv.Atoi(strings.Split(matches[3], ",")[1])
			if err != nil {
				return nil, &DecodeError{
					Cmd:  ".B",
					Line: string(line),
					Msg:  "Cannot parse lock state",
				}
			}

			cmd.Locked = locked
		}
		return cmd, nil
	}
	return nil, &DecodeError{
		Cmd:  ".B",
		Line: string(line),
		Msg:  "Invalid request",
	}
}

func DecodeDotF(line []byte) (*DotFCmd, error) {
	// .F1, .F001, .F01,... .F032, .F32
	ok := dotFRegexp.Match(line)
	if !ok {
		return nil, &DecodeError{
			Cmd:  ".F",
			Line: string(line),
			Msg:  "Invalid format",
		}
	}
	matches := dotFRegexp.FindStringSubmatch(string(line))
	if matches != nil {
		val, err := strconv.Atoi(matches[0])
		if err != nil {
			return nil, &DecodeError{
				Cmd:  ".F",
				Line: string(line),
				Msg:  "Cannot parse salvo",
			}
		}
		if val > 32 || val < 1 {
			return nil, &DecodeError{
				Cmd:  ".F",
				Line: string(line),
				Msg:  "Salvo out of range",
			}
		}
		return &DotFCmd{
			Salvo: val,
		}, nil
	}

	return nil, &DecodeError{
		Cmd:  ".F",
		Line: string(line),
		Msg:  "Invalid request",
	}
}

func DecodeDotI(line []byte) (*DotICmd, error) {
	// .IV1, .IA1
	ok := dotIRegexp.Match(line)
	if !ok {
		return nil, &DecodeError{
			Cmd:  ".I",
			Line: string(line),
			Msg:  "Invalid format",
		}
	}
	matches := dotIRegexp.FindStringSubmatch(string(line))
	if matches != nil {
		level := matches[1]
		dst, err := strconv.Atoi(matches[2])
		if err != nil {
			return nil, &DecodeError{
				Cmd:  ".I",
				Line: string(line),
				Msg:  "Cannot parse dst",
			}
		}
		if !(dst > 0) {
			return nil, &DecodeError{
				Cmd:  ".I",
				Line: string(line),
				Msg:  "Dst must be > 0",
			}
		}
		return &DotICmd{
			Level: level,
			Dst:   dst,
		}, nil
	}

	return nil, &DecodeError{
		Cmd:  ".I",
		Line: string(line),
		Msg:  "Invalid request",
	}
}

func DecodeDotL(line []byte) (*DotLCmd, error) {
	// .LV1,- OR .LV1,1
	ok := dotLRegexp.Match(line)
	if !ok {
		return nil, &DecodeError{
			Cmd:  ".L",
			Line: string(line),
			Msg:  "Invalid format",
		}
	}
	matches := dotLRegexp.FindStringSubmatch(string(line))
	if matches != nil {
		level := matches[1]
		dstRange := strings.Split(matches[2], ",")
		if !(len(dstRange) > 1) {
			return nil, &DecodeError{
				Cmd:  ".L",
				Line: string(line),
				Msg:  "Invalid command format",
			}

		}
		dstStr := dstRange[0]

		dst, err := strconv.Atoi(dstStr)
		if err != nil {
			return nil, &DecodeError{
				Cmd:  ".L",
				Line: string(line),
				Msg:  "Cannot parse dst",
			}
		}

		if !(dst > 0) {
			return nil, &DecodeError{
				Cmd:  ".L",
				Line: string(line),
				Msg:  "Dst must be > 0",
			}
		}

		cmd := &DotLCmd{
			Level: level,
			Dst:   dst,
		}

		if matches[3] != "" {
			if strings.Contains(matches[2], "-") {
				return nil, &DecodeError{
					Cmd:  ".L",
					Line: string(line),
					Msg:  "Invalid request",
				}
			}
			src, err := strconv.Atoi(matches[3])
			if err != nil {
				return nil, &DecodeError{
					Cmd:  ".L",
					Line: string(line),
					Msg:  "Cannot parse src",
				}
			}
			cmd.Src = src
		}
		return cmd, nil
	}

	return nil, &DecodeError{
		Cmd:  ".L",
		Line: string(line),
		Msg:  "Invalid request",
	}
}

func DecodeDotR(line []byte) (*DotRCmd, error) {
	// .RD1/.RE1, RS1/.RT1, .RLV/.RMV, .RA[D/S/L/E/T/M]...
	ok := dotRRegexp.Match(line)
	if !ok {
		return nil, &DecodeError{
			Cmd:  ".R",
			Line: string(line),
			Msg:  "Invalid format",
		}
	}
	matches := dotRRegexp.FindStringSubmatch(string(line))
	if matches != nil {
		cmdType := matches[1]
		destSrcLevel := matches[2]

		cmd := &DotRCmd{
			Type: cmdType,
		}

		if destSrcLevel == "" {
			// not the comma response syntax
			switch cmdType {
			// could be a command
			case "D", "S", "E", "T":
				// the 3rd match should be a dest or source index
				dstSrc, err := strconv.Atoi(matches[3])
				if err != nil {
					return nil, &DecodeError{
						Cmd:  ".R",
						Line: string(line),
						Msg:  "Invalid dest/src in command",
					}
				}
				cmd.DstSrc = dstSrc
				if !(dstSrc > 0) {
					return nil, &DecodeError{
						Cmd:  ".R",
						Line: string(line),
						Msg:  "Dst/Src specifier must be > 0",
					}
				}

			case "L", "M":
				level := matches[3]
				cmd.Level = level
			default:
				mnemonic := matches[3]
				cmd.Mnemonic = mnemonic
			}
		} else {
			dsl := strings.Split(destSrcLevel, ",")
			if len(dsl) == 1 {
				return nil, &DecodeError{
					Cmd:  ".R",
					Line: string(line),
					Msg:  "Invalid dest/src/level in response",
				}
			}
			dstsrc, err := strconv.Atoi(dsl[0])
			if err != nil {
				// would be a level response
				level := dsl[0]
				cmd.Level = level
			} else {
				cmd.DstSrc = dstsrc
			}
			mnemonic := matches[3]
			cmd.Mnemonic = mnemonic
		}
		return cmd, nil
	}

	return nil, &DecodeError{
		Cmd:  ".R",
		Line: string(line),
		Msg:  "Invalid request",
	}
}

func DecodeDotW(line []byte) (*DotWCmd, error) {
	ok := dotWRegexp.Match(line)
	if !ok {
		return nil, &DecodeError{
			Cmd:  ".W",
			Line: string(line),
			Msg:  "Invalid format",
		}
	}
	matches := dotWRegexp.FindStringSubmatch(string(line))
	if matches != nil {
		cmdType := matches[1]
		destSrcLevel := matches[2]
		mnemonic := matches[3]

		cmd := &DotWCmd{
			Type:     cmdType,
			Mnemonic: mnemonic,
		}

		dstsrc, err := strconv.Atoi(destSrcLevel)
		if err != nil {
			// would be a level command
			level := destSrcLevel
			cmd.Level = level
		} else {
			cmd.DstSrc = dstsrc
		}
		return cmd, nil
	}

	return nil, &DecodeError{
		Cmd:  ".W",
		Line: string(line),
		Msg:  "Invalid request",
	}
}

func DecodeDotQ(line []byte) (*DotQCmd, error) {
	ok := dotQRegexp.Match(line)
	if !ok {
		return nil, &DecodeError{
			Cmd:  ".Q",
			Line: string(line),
			Msg:  "Invalid request",
		}
	}
	matches := dotQRegexp.FindStringSubmatch(string(line))
	if matches != nil {
		cmdType := matches[1]
		salvoStr := matches[2]
		level := matches[3]
		destStr := matches[4]
		srcStr := matches[5]
		timeStr := matches[6]

		salvo, err := strconv.Atoi(salvoStr)
		if err != nil {
			return nil, &DecodeError{
				Cmd:  ".Q",
				Line: string(line),
				Msg:  "Cannot infer salvo number",
			}

		}

		if !(salvo >= 0) {
			return nil, &DecodeError{
				Cmd:  ".Q",
				Line: string(line),
				Msg:  "Salvo number must be >=0",
			}
		}

		cmd := &DotQCmd{
			Type:  cmdType,
			Salvo: salvo,
		}

		switch cmdType {
		case "S":
			dest, err := strconv.Atoi(destStr)
			if err != nil {
				return nil, &DecodeError{
					Cmd:  ".Q",
					Line: string(line),
					Msg:  "Cannot infer destination integer",
				}

			}
			src, err := strconv.Atoi(srcStr)
			if err != nil {
				return nil, &DecodeError{
					Cmd:  ".Q",
					Line: string(line),
					Msg:  "Cannot infer source integer",
				}

			}
			if level == "" || (!(dest > 0) || !(src > 0)) {
				return nil, &DecodeError{
					Cmd:  ".Q",
					Line: string(line),
					Msg:  "Invalid request",
				}
			}
			cmd.Level = level
			cmd.Dest = dest
			cmd.Src = src
			return cmd, nil
		case "F":
			if timeStr == "" {
				// just a default .QF
				return cmd, nil
			}
			timeSlice := strings.Split(timeStr, ":")
			if len(timeSlice) > 1 {
				hhs := timeSlice[1]
				mms := timeSlice[2]
				sss := timeSlice[3]
				ffs := timeSlice[4]

				hh, err := strconv.Atoi(hhs)
				if err != nil {
					return nil, &DecodeError{
						Cmd:  ".Q",
						Line: string(line),
						Msg:  "Cannot infer hh in timestamp",
					}
				}
				mm, err := strconv.Atoi(mms)
				if err != nil {
					return nil, &DecodeError{
						Cmd:  ".Q",
						Line: string(line),
						Msg:  "Cannot infer mm in timestamp",
					}
				}
				ss, err := strconv.Atoi(sss)
				if err != nil {
					return nil, &DecodeError{
						Cmd:  ".Q",
						Line: string(line),
						Msg:  "Cannot infer ss in timestamp",
					}
				}
				ff, err := strconv.Atoi(ffs)
				if err != nil {
					return nil, &DecodeError{
						Cmd:  ".Q",
						Line: string(line),
						Msg:  "Cannot infer ff in timestamp",
					}
				}

				ts := &DotQFTime{
					Hours:   hh,
					Minutes: mm,
					Seconds: ss,
					Frames:  ff,
				}

				cmd.FTime = ts

				return cmd, nil

			} else {
				return nil, &DecodeError{
					Cmd:  ".Q",
					Line: string(line),
					Msg:  "Invalid request",
				}
			}
		case "L":
			if len(matches) > 7 && matches[7] != "" {
				val, err := strconv.Atoi(matches[7])
				if err != nil {
					return nil, &DecodeError{
						Cmd:  ".Q",
						Line: string(line),
						Msg:  "Cannot parse level count",
					}
				}
				cmd.Count = val
				cmd.Response = true
			}
			return cmd, nil
		default:
			return cmd, nil
		}
	}

	return nil, &DecodeError{
		Cmd:  ".Q",
		Line: string(line),
		Msg:  "Invalid request",
	}
}
