package quartz

type Decoder[T any] func(cmd string) T

type DotSCmd struct {
	levels []string
	dst    string
	src    string
}

func DecodeDotS[T DotSCmd](line []byte) (*DotSCmd, error) {
	// .SVABCD1,1
	return &DotSCmd{}, nil
}
