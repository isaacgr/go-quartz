package goquartz

import (
	"reflect"
	"testing"
)

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
		//{"valid dest level add param", []byte("V1+A1+B2,1"), false},
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
				Locked: 0,
			},
			false,
		},
		{
			"valid acknowledge lock 1",
			[]byte("A1,1"),
			&DstLockCmd{
				Dst:    1,
				Type:   "A",
				Locked: 1,
			},
			false,
		},
		{
			"valid acknowledge lock 255",
			[]byte("A1,255"),
			&DstLockCmd{
				Dst:    1,
				Type:   "A",
				Locked: 255,
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
					t.Errorf("DecodeDotB(%q) unexpected error = %v", tt.input, err)
				}
			} else {
				if tt.wantErr {
					t.Errorf("DecodeDotB(%q) expected error, but got none", tt.input)
				} else if !reflect.DeepEqual(cmd, tt.wantCmd) {
					t.Errorf("DecodeDotB(%q) got = %+v, want %+v", tt.input, cmd, tt.wantCmd)
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
				Src:   1,
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
					t.Errorf("DecodeDotL(%q) unexpected error = %v", tt.input, err)
				}
			} else {
				if tt.wantErr {
					t.Errorf("DecodeDotL(%q) expected error, but got none", tt.input)
				} else if !reflect.DeepEqual(cmd, tt.wantCmd) {
					t.Errorf("DecodeDotL(%q) got = %+v, want %+v", tt.input, cmd, tt.wantCmd)
				}
			}
		})
	}
}

func TestDecodeDotR(t *testing.T) {
	tests := []struct {
		name    string
		input   []byte
		wantCmd *DotRCmd
		wantErr bool
	}{
		{
			"Read destination index",
			[]byte("D1"),
			&DotRCmd{Type: "D", DstSrc: 1},
			false,
		},
		{
			"Read source index",
			[]byte("S1"),
			&DotRCmd{Type: "S", DstSrc: 1},
			false,
		},
		{
			"Read level",
			[]byte("LV"),
			&DotRCmd{Type: "L", Level: "V"},
			false,
		},
		{
			"Read protect dest",
			[]byte("E1"),
			&DotRCmd{Type: "E", DstSrc: 1},
			false,
		},
		{
			"Read protect src",
			[]byte("T1"),
			&DotRCmd{Type: "T", DstSrc: 1},
			false,
		},
		{
			"Read matrix",
			[]byte("MV"),
			&DotRCmd{Type: "M", Level: "V"},
			false,
		},
		{
			"Query destination name",
			[]byte("ADsomename"),
			&DotRCmd{Type: "AD", Mnemonic: "somename"},
			false,
		},
		{
			"Query source name",
			[]byte("ASsomename"),
			&DotRCmd{Type: "AS", Mnemonic: "somename"},
			false,
		},
		{
			"Query level name",
			[]byte("ALsomename"),
			&DotRCmd{Type: "AL", Mnemonic: "somename"},
			false,
		},
		{
			"Query protect dest name",
			[]byte("AEsomename"),
			&DotRCmd{Type: "AE", Mnemonic: "somename"},
			false,
		},

		{
			"Associate destination index with name",
			[]byte("AD1,somename"),
			&DotRCmd{Type: "AD", DstSrc: 1, Mnemonic: "somename"},
			false,
		},
		{
			"Associate source index with name",
			[]byte("AS1,somename"),
			&DotRCmd{Type: "AS", DstSrc: 1, Mnemonic: "somename"},
			false,
		},
		{
			"Associate level name with name",
			[]byte("ALV,somename"),
			&DotRCmd{Type: "AL", Level: "V", Mnemonic: "somename"},
			false,
		},
		{
			"Associate protect dest index with name",
			[]byte("AE1,somename"),
			&DotRCmd{Type: "AE", DstSrc: 1, Mnemonic: "somename"},
			false,
		},
		{
			"Associate matrix name with name",
			[]byte("AMV,somename"),
			&DotRCmd{Type: "AM", Level: "V", Mnemonic: "somename"},
			false,
		},

		{
			"Mnemonic with mixed casing",
			[]byte("ADSomeone"),
			&DotRCmd{Type: "AD", Mnemonic: "Someone"},
			false,
		},
		{
			"Mnemonic with special characters",
			[]byte("AD !@#$%%^&*()?.[';/.,?><;'"),
			&DotRCmd{Type: "AD", Mnemonic: " !@#$%%^&*()?.[';/.,?><;'"},
			false,
		},

		{
			"Invalid prefix",
			[]byte("XZ1"),
			nil,
			true,
		},
		{
			"Empty input",
			[]byte(""),
			nil,
			true,
		},
		{
			"Invalid empty index",
			[]byte("D0"),
			nil,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, err := DecodeDotR(tt.input)
			if err != nil {
				if !tt.wantErr {
					t.Errorf("DecodeDotR(%q) unexpected error = %v", tt.input, err)
				}
			} else {
				if tt.wantErr {
					t.Errorf("DecodeDotR(%q) expected error, but got none", tt.input)
				} else if !reflect.DeepEqual(cmd, tt.wantCmd) {
					t.Errorf("DecodeDotR(%q) got = %+v, want %+v", tt.input, cmd, tt.wantCmd)
				}
			}
		})
	}
}

func TestDecodeDotW(t *testing.T) {
	tests := []struct {
		name    string
		input   []byte
		wantCmd *DotWCmd
		wantErr bool
	}{
		{
			"Write destination mnemonic",
			[]byte("D1,somename"),
			&DotWCmd{Type: "D", DstSrc: 1, Mnemonic: "somename"},
			false,
		},
		{
			"Write source mnemonic",
			[]byte("S1,somename"),
			&DotWCmd{Type: "S", DstSrc: 1, Mnemonic: "somename"},
			false,
		},
		{
			"Write level mnemonic",
			[]byte("LV,somename"),
			&DotWCmd{Type: "L", Level: "V", Mnemonic: "somename"},
			false,
		},
		{
			"Write empty mnemonic to clear",
			[]byte("D1,"),
			&DotWCmd{Type: "D", DstSrc: 1, Mnemonic: ""},
			false,
		},
		{
			"Missing comma and mnemonic",
			[]byte("D1"),
			nil,
			true,
		},
		{
			"Missing index/level",
			[]byte("Dsomename"),
			nil,
			true,
		},
		{
			"Invalid prefix",
			[]byte("X1,somename"),
			nil,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, err := DecodeDotW(tt.input)
			if err != nil {
				if !tt.wantErr {
					t.Errorf("DecodeDotW(%q) unexpected error = %v", tt.input, err)
				}
			} else {
				if tt.wantErr {
					t.Errorf("DecodeDotW(%q) expected error, but got none", tt.input)
				} else if !reflect.DeepEqual(cmd, tt.wantCmd) {
					t.Errorf("DecodeDotW(%q) got = %+v, want %+v", tt.input, cmd, tt.wantCmd)
				}
			}
		})
	}
}

func TestDecodeDotQ(t *testing.T) {
	tests := []struct {
		name    string
		input   []byte
		wantCmd *DotQCmd
		wantErr bool
	}{
		{
			"Create salvo",
			[]byte("C1"),
			&DotQCmd{Type: "C", Salvo: 1},
			false,
		},
		{
			"Empty salvo",
			[]byte("R1"),
			&DotQCmd{Type: "R", Salvo: 1},
			false,
		},
		{
			"Destroy salvo",
			[]byte("D1"),
			&DotQCmd{Type: "D", Salvo: 1},
			false,
		},
		{
			"List salvo",
			[]byte("L1"),
			&DotQCmd{Type: "L", Salvo: 1},
			false,
		},
		{
			"List salvo response with elements",
			[]byte("L1,12"),
			&DotQCmd{Type: "L", Salvo: 1, Count: 12, Response: true},
			false,
		},
		{
			"List salvo response with zero elements",
			[]byte("L0,0"),
			&DotQCmd{Type: "L", Salvo: 0, Count: 0, Response: true},
			false,
		},
		{
			"Fire salvo time default",
			[]byte("F1"),
			&DotQCmd{Type: "F", Salvo: 1},
			false,
		},
		{
			"Load salvo",
			[]byte("S1V1,2"),
			&DotQCmd{Type: "S", Salvo: 1, Level: "V", Dest: 1, Src: 2},
			false,
		},
		{
			"Fire salvo with timestamp",
			[]byte("F1T1:12:34:56:00"),
			&DotQCmd{
				Type:  "F",
				Salvo: 1,
				FTime: &DotQFTime{
					Hours:   12,
					Minutes: 34,
					Seconds: 56,
					Frames:  0,
				},
			},
			false,
		},
		// Invalid cases
		{
			"Missing salvo number",
			[]byte("C"),
			nil,
			true,
		},
		{
			"Negative salvo number",
			[]byte("C-1"),
			nil,
			true,
		},
		{
			"S command missing level/dest/src",
			[]byte("S1"),
			nil,
			true,
		},
		{
			"S command missing src",
			[]byte("S1V1,"),
			nil,
			true,
		},
		{
			"S command missing level",
			[]byte("S11,2"),
			nil,
			true,
		},
		{
			"F command invalid timestamp format",
			[]byte("F1T1:12:34:56"),
			nil,
			true,
		},
		{
			"Invalid command prefix",
			[]byte("X1"),
			nil,
			true,
		},
		{
			"Empty input",
			[]byte(""),
			nil,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, err := DecodeDotQ(tt.input)
			if err != nil {
				if !tt.wantErr {
					t.Errorf("DecodeDotQ(%q) unexpected error = %v", tt.input, err)
				}
			} else {
				if tt.wantErr {
					t.Errorf("DecodeDotQ(%q) expected error, but got none", tt.input)
				} else if !reflect.DeepEqual(cmd, tt.wantCmd) {
					t.Errorf("DecodeDotQ(%q) got = %+v, want %+v", tt.input, cmd, tt.wantCmd)
				}
			}
		})
	}
}
