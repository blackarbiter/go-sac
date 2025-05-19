package domain_test

import (
	"github.com/blackarbiter/go-sac/pkg/domain"
	"testing"
)

func TestSecurityMaterialTypeString(t *testing.T) {
	tests := []struct {
		input    domain.SecurityMaterialType
		expected string
	}{
		{domain.SecurityMaterialTypeSensitiveWords, "SensitiveWords"},
		{domain.SecurityMaterialTypeThreatModel, "ThreatModel"},
		{domain.SecurityMaterialTypeSecuritySpec, "SecuritySpec"},
		{domain.SecurityMaterialTypeToolkit, "Toolkit"},
		{domain.SecurityMaterialTypeScanRule, "ScanRule"},
		{domain.SecurityMaterialTypeSecurityStandard, "SecurityStandard"},
		{domain.SecurityMaterialTypeComplianceDoc, "ComplianceDoc"},
		{domain.SecurityMaterialTypeUnknown, "Unknown"},
		{100, "Unknown"},
	}

	for _, tt := range tests {
		result := tt.input.String()
		if result != tt.expected {
			t.Errorf("For %d expected %s but got %s", tt.input, tt.expected, result)
		}
	}
}

func TestParseSecurityMaterialType(t *testing.T) {
	tests := []struct {
		input    string
		expected domain.SecurityMaterialType
		hasError bool
	}{
		{"SensitiveWords", domain.SecurityMaterialTypeSensitiveWords, false},
		{"ThreatModel", domain.SecurityMaterialTypeThreatModel, false},
		{"SecuritySpec", domain.SecurityMaterialTypeSecuritySpec, false},
		{"Toolkit", domain.SecurityMaterialTypeToolkit, false},
		{"ScanRule", domain.SecurityMaterialTypeScanRule, false},
		{"SecurityStandard", domain.SecurityMaterialTypeSecurityStandard, false},
		{"ComplianceDoc", domain.SecurityMaterialTypeComplianceDoc, false},
		{"InvalidType", domain.SecurityMaterialTypeUnknown, true},
		{"", domain.SecurityMaterialTypeUnknown, true},
	}

	for _, tt := range tests {
		result, err := domain.ParseSecurityMaterialType(tt.input)
		if (err != nil) != tt.hasError {
			t.Errorf("ParseSecurityMaterialType(%s) error = %v, wantErr %v", tt.input, err, tt.hasError)
			continue
		}
		if result != tt.expected {
			t.Errorf("ParseSecurityMaterialType(%s) = %v, want %v", tt.input, result, tt.expected)
		}
	}
}
