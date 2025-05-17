package validator_test

import (
	"testing"

	"github.com/blackarbiter/go-sac/pkg/utils/validator"
)

func TestSQLInjectionCheck(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"SELECT * FROM users", true},
		{"' OR '1'='1'--", true},
		{"normal input", false},
		{"name; DROP TABLE users;", true},
		{"UPDATE users SET password='123' WHERE id=1", true},
	}

	for _, tt := range tests {
		got := validator.CheckSQLInjection(tt.input)
		if got != tt.want {
			t.Errorf("Input: %q, Got: %v, Want: %v", tt.input, got, tt.want)
		}
	}
}
