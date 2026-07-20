package fp

import (
	"testing"

	"github.com/shopspring/decimal"
)

func TestParseBalanceAmount(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string // decimal.String() канонический
		wantErr bool
	}{
		{"rub_zero", "0 ₽", "0", false},
		{"rub_thousands", "1 234,56 ₽", "1234.56", false},
		{"rub_thousands_nbsp", "1\u00A0234,56 ₽", "1234.56", false},
		{"usd_dot", "99.50 $", "99.5", false}, // decimal каноникализует 99.50 → 99.5
		{"usd_thousands_comma", "$1,234.56", "1234.56", false},
		{"plain_comma", "12345,67", "12345.67", false},
		{"plain_dot", "12345.67", "12345.67", false},
		{"empty", "", "0", true},
		{"not_number", "abc", "0", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseBalanceAmount(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("input %q: want error, got nil (result %s)", tt.input, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("input %q: unexpected error: %v", tt.input, err)
			}
			want, _ := decimal.NewFromString(tt.want)
			if !got.Equal(want) {
				t.Errorf("input %q: got %s, want %s", tt.input, got, want)
			}
		})
	}
}
