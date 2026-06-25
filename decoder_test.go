package goquartz

import "testing"

func TestDecodeDotS(t *testing.T) {
	tests := []struct {
		name    string
		input   []byte
		wantErr bool
	}{
		{"valid single digit", []byte("V1,1"), false},
		{"valid multi digit", []byte("VAB001,001"), false},
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
