package goquartz

import "testing"

func TestDecodeDotE(t *testing.T) {
	tests := []struct {
		name    string
		input   []byte
		wantErr bool
	}{
		{"valid", []byte("E"), false},
		{"extra info", []byte("E - Invalid route"), true}, // TODO: This may change
		{"starts with another letter", []byte("EE"), true},
		{"empty", []byte(""), true},
		{"lowercase", []byte("e"), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := DecodeDotE(tt.input)
			if err != nil {
				if !tt.wantErr {
					t.Errorf(
						"DecodeDotE(%q) error = %v, wantErr %v",
						tt.input,
						err,
						tt.wantErr,
					)
				}
			}
		})
	}
}

func TestDecodeDotA(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		wantResp []DotAResp
		wantErr  bool
	}{
		{
			"valid level,dst,src",
			[]byte("V1,1"),
			[]DotAResp{
				DotAResp{
					Level: "V",
					Src:   1,
					Dst:   1,
				},
			},
			false,
		},
		{
			"valid multi level,dst,src",
			[]byte("V1,1A1,1B1,2"),
			[]DotAResp{
				DotAResp{
					Level: "V",
					Src:   1,
					Dst:   1,
				},
				DotAResp{
					Level: "A",
					Src:   1,
					Dst:   1,
				},
				DotAResp{
					Level: "B",
					Src:   2,
					Dst:   1,
				},
			},
			false,
		},
		{
			"invalid level,dst",
			[]byte("V1,"),
			nil,
			true,
		},
		{
			"invalid multiple levels",
			[]byte("VABC1,1"),
			nil,
			true,
		},
		{
			"extra info",
			// TODO: This may change
			[]byte("A - Invalid route"),
			nil,
			true,
		},
		{
			"starts with another letter",
			[]byte("AA"),
			nil,
			true,
		},
		{
			"empty",
			[]byte(""),
			nil,
			true,
		},
		{
			"lowercase",
			[]byte("a"),
			nil,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := DecodeDotA(tt.input)
			if err != nil {
				if !tt.wantErr {
					t.Errorf(
						"DecodeDotA(%q) error = %v, wantErr %v",
						tt.input,
						err,
						tt.wantErr,
					)
				}
			} else {
				if tt.wantErr {
					t.Errorf(
						"DecodeDotA(%q) error = %v, wantErr %v",
						tt.input,
						err,
						tt.wantErr,
					)
				} else {
					for i, r := range tt.wantResp {
						if r.Level != resp[i].Level {
							t.Errorf(
								"Incorrect level received. got=%v, want=%v",
								resp,
								r,
							)
						}
						if r.Dst != resp[i].Dst {
							t.Errorf(
								"Incorrect dst received. got=%v, want=%v",
								resp,
								r,
							)
						}
						if r.Src != resp[i].Src {
							t.Errorf(
								"Incorrect src received. got=%v, want=%v",
								resp,
								r,
							)
						}
					}
				}
			}
		})
	}
}

func TestDecodeDotS(t *testing.T) {
	tests := []struct {
		name    string
		input   []byte
		wantErr bool
	}{
		{"valid single digit", []byte("V1,1"), false},
		{"valid multi-level", []byte("VABCDEFGHIJKLMNOP1,1"), false},
		{"valid multi digit", []byte("VAB001,001"), false},
		{"invalid multi-level", []byte("VABCDEFGHIJKLMNOPQ1,1"), true},
		{"no digits", []byte("V,"), true},
		{"missing comma", []byte("V11"), true},
		{"starts with digit", []byte("1V,1"), true},
		{"empty", []byte(""), true},
		{"lowercase", []byte("v1,1"), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := DecodeDotS(tt.input)
			if err != nil {
				if !tt.wantErr {
					t.Errorf(
						"DecodeDotS(%q) error = %v, wantErr %v",
						tt.input,
						err,
						tt.wantErr,
					)
				}
			}
		})
	}
}

func TestDecodeDotM(t *testing.T) {
	tests := []struct {
		name    string
		input   []byte
		wantErr bool
	}{
		{"valid single digit", []byte("V1,1"), false},
		{"valid multi-level", []byte("VA1,1"), false},
		{"valid multi digit", []byte("VAB001,001"), false},
		{"valid multi route", []byte("V1,1,A1,1,B1,1"), false},
		{"valid dest range", []byte("VA1-5,1"), false},
		{"valid dest range multi digit", []byte("VA001-005,001"), false},
		{"valid src range", []byte("VA1-5,10-14"), false},
		{"valid src range multi digit", []byte("VA001-005,010-014"), false},
		{"valid dest level add param", []byte("V1+A1+B2,1"), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := DecodeDotM(tt.input)
			if err != nil {
				if !tt.wantErr {
					t.Errorf(
						"DecodeDotM(%q) error = %v, wantErr %v",
						tt.input,
						err,
						tt.wantErr,
					)
				}
			}
		})
	}
}

func makeIntPointer(v int) *int {
	return &v
}

func TestDecodeDotB(t *testing.T) {

	tests := []struct {
		name    string
		input   []byte
		wantCmd *DstLockCmd
		wantErr bool
	}{
		{
			"valid lock",
			[]byte("L1"),
			&DstLockCmd{
				Dst:  1,
				Type: "L",
			},
			false,
		},
		{
			"valid unlock",
			[]byte("U1"),
			&DstLockCmd{
				Dst:  1,
				Type: "U",
			},
			false,
		},
		{
			"valid inspect",
			[]byte("I1"),
			&DstLockCmd{
				Dst:  1,
				Type: "I",
			},
			false,
		},
		{
			"invalid command type",
			[]byte("X1"),
			nil,
			true,
		},
		{
			"valid acknowledge no lock",
			[]byte("A1,0"),
			&DstLockCmd{
				Dst:    1,
				Type:   "A",
				Locked: makeIntPointer(0),
			},
			false,
		},
		{
			"valid acknowledge lock 1",
			[]byte("A1,1"),
			&DstLockCmd{
				Dst:    1,
				Type:   "A",
				Locked: makeIntPointer(1),
			},
			false,
		},
		{
			"valid acknowledge lock 255",
			[]byte("A1,255"),
			&DstLockCmd{
				Dst:    1,
				Type:   "A",
				Locked: makeIntPointer(255),
			},
			false,
		},
		{
			"invalid acknowledge",
			[]byte("A"),
			nil,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, err := DecodeDotB(tt.input)
			if err != nil {
				if !tt.wantErr {
					t.Errorf(
						"DecodeDotB(%q) error = %v, wantErr %v",
						tt.input,
						err,
						tt.wantErr,
					)
				}
			} else {
				if cmd.Dst != tt.wantCmd.Dst {
					t.Errorf(
						"Incorrect dst received. got=%v, want=%v",
						cmd,
						tt.wantCmd,
					)
				}
				if cmd.Locked != nil && tt.wantCmd != nil {
					if *cmd.Locked != *tt.wantCmd.Locked {
						t.Errorf(
							"Incorrect lock state received. got=%v, want=%v",
							cmd,
							tt.wantCmd,
						)
					}
				}
				if cmd.Type != tt.wantCmd.Type {
					t.Errorf(
						"Incorrect lock type received. got=%v, want=%v",
						cmd,
						tt.wantCmd,
					)
				}
			}
		})
	}
}

func TestDecodeDotF(t *testing.T) {

	tests := []struct {
		name    string
		input   []byte
		wantCmd *DotFCmd
		wantErr bool
	}{
		{
			"valid lock 1",
			[]byte("1"),
			&DotFCmd{
				Salvo: 1,
			},
			false,
		},
		{
			"valid lock 01",
			[]byte("01"),
			&DotFCmd{
				Salvo: 1,
			},
			false,
		},
		{
			"valid lock 001",
			[]byte("001"),
			&DotFCmd{
				Salvo: 1,
			},
			false,
		},
		{
			"valid lock 10",
			[]byte("10"),
			&DotFCmd{
				Salvo: 10,
			},
			false,
		},
		{
			"valid lock 32",
			[]byte("32"),
			&DotFCmd{
				Salvo: 32,
			},
			false,
		},
		{
			"valid lock 032",
			[]byte("032"),
			&DotFCmd{
				Salvo: 32,
			},
			false,
		},
		{
			"invalid lock 0",
			[]byte("0"),
			nil,
			true,
		},
		{
			"invalid lock 000",
			[]byte("000"),
			nil,
			true,
		},
		{
			"invalid lock x",
			[]byte("x"),
			nil,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, err := DecodeDotF(tt.input)
			if err != nil {
				if !tt.wantErr {
					t.Errorf(
						"DecodeDotF(%q) error = %v, wantErr %v",
						tt.input,
						err,
						tt.wantErr,
					)
				}
			} else {
				if cmd.Salvo != tt.wantCmd.Salvo {
					t.Errorf(
						"Incorrect salvo received. got=%v, want=%v",
						cmd,
						tt.wantCmd,
					)
				}
			}
		})
	}
}

func TestDecodeDotI(t *testing.T) {

	tests := []struct {
		name    string
		input   []byte
		wantCmd *DotICmd
		wantErr bool
	}{
		{
			"valid V1",
			[]byte("V1"),
			&DotICmd{
				Level: "V",
				Dst:   1,
			},
			false,
		},
		{
			"valid V01",
			[]byte("V01"),
			&DotICmd{
				Level: "V",
				Dst:   1,
			},
			false,
		},
		{
			"invalid V0",
			[]byte("V0"),
			nil,
			true,
		},
		{
			"invalid v1",
			[]byte("v1"),
			nil,
			true,
		},
		{
			"invalid VA1",
			[]byte("VA1"),
			nil,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, err := DecodeDotI(tt.input)
			if err != nil {
				if !tt.wantErr {
					t.Errorf(
						"DecodeDotI(%q) error = %v, wantErr %v",
						tt.input,
						err,
						tt.wantErr,
					)
				}
			} else {
				if cmd.Level != tt.wantCmd.Level {
					t.Errorf(
						"Incorrect level received. got=%v, want=%v",
						cmd,
						tt.wantCmd,
					)
				}
				if cmd.Dst != tt.wantCmd.Dst {
					t.Errorf(
						"Incorrect dst received. got=%v, want=%v",
						cmd,
						tt.wantCmd,
					)
				}
			}
		})
	}
}

func TestDecodeDotL(t *testing.T) {

	tests := []struct {
		name    string
		input   []byte
		wantCmd *DotLCmd
		wantErr bool
	}{
		{
			"valid V1,-",
			[]byte("V1,-"),
			&DotLCmd{
				Level: "V",
				Dst:   1,
			},
			false,
		},
		{
			"valid V01,-",
			[]byte("V01,-"),
			&DotLCmd{
				Level: "V",
				Dst:   1,
			},
			false,
		},
		{
			"valid V1,1",
			[]byte("V1,1"),
			&DotLCmd{
				Level: "V",
				Dst:   1,
				Src:   makeIntPointer(1),
			},
			false,
		},
		{
			"invalid V1",
			[]byte("V1"),
			nil,
			true,
		},
		{
			"invalid V1,-1",
			[]byte("V1,-1"),
			nil,
			true,
		},
		{
			"invalid V1,",
			[]byte("V1,-1"),
			nil,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, err := DecodeDotL(tt.input)
			if err != nil {
				if !tt.wantErr {
					t.Errorf(
						"DecodeDotL(%q) error = %v, wantErr %v",
						tt.input,
						err,
						tt.wantErr,
					)
				}
			} else {
				if tt.wantErr {
					t.Errorf(
						"DecodeDotL(%q) error = %v, wantErr %v",
						tt.input,
						err,
						tt.wantErr,
					)
				} else {
					if cmd.Level != tt.wantCmd.Level {
						t.Errorf(
							"Incorrect level received. got=%v, want=%v",
							cmd,
							tt.wantCmd,
						)
					}
					if cmd.Dst != tt.wantCmd.Dst {
						t.Errorf(
							"Incorrect dst received. got=%v, want=%v",
							cmd,
							tt.wantCmd,
						)
					}

				}
			}
		})
	}
}
