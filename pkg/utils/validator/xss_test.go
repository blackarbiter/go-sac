package validator_test

import (
	"testing"

	"github.com/blackarbiter/go-sac/pkg/utils/validator"
)

func TestXSSSanitization(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"<script>alert('xss')</script>", ""},
		{"<img src=x onerror=alert(1)>", "<img src=\"x\">"},
		{"<a href=\"javascript:alert(1)\">click</a>", "click"},
	}

	for _, tt := range tests {
		got := validator.SanitizeXSSRelaxed(tt.input)
		if got != tt.want {
			t.Errorf("Input: %q\n Got: %q\n Want: %q", tt.input, got, tt.want)
		}
	}
}
