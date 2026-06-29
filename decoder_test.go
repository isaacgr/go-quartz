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
		name    string
		input   []byte
		wantErr bool
	}{
		{"valid", []byte("A"), false},
		{"extra info", []byte("A - Invalid route"), true}, // TODO: This may change
		{"starts with another letter", []byte("AA"), true},
		{"empty", []byte(""), true},
		{"lowercase", []byte("a"), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := DecodeDotA(tt.input)
			if err != nil {
				if !tt.wantErr {
					t.Errorf(
						"DecodeDotA(%q) error = %v, wantErr %v",
						tt.input,
						err,
						tt.wantErr,
					)
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
