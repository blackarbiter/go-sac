package domain_test

import (
	"github.com/blackarbiter/go-sac/pkg/domain"
	"testing"
)

func TestScanTypeString(t *testing.T) {
	tests := []struct {
		input    domain.ScanType
		expected string
	}{
		{domain.ScanTypeRequirementAnalysis, "RequirementAnalysis"},
		{domain.ScanTypeThreatModeling, "ThreatModeling"},
		{domain.ScanTypeSecuritySpecCheck, "SecuritySpecCheck"},
		{domain.ScanTypeStaticCodeAnalysis, "StaticCodeAnalysis"},
		{domain.ScanTypeContainerImageScan, "ContainerImageScan"},
		{domain.ScanTypeHostSecurityCheck, "HostSecurityCheck"},
		{domain.ScanTypeBlackBoxTesting, "BlackBoxTesting"},
		{domain.ScanTypePortScanning, "PortScanning"},
		{domain.ScanTypeDast, "DAST"},
		{domain.ScanTypeSca, "SCA"},
		{domain.ScanTypeSecretsDetection, "SecretsDetection"},
		{domain.ScanTypeComplianceAudit, "ComplianceAudit"},
		{domain.ScanTypeUnknown, "Unknown"},
		{100, "Unknown"},
	}

	for _, tt := range tests {
		result := tt.input.String()
		if result != tt.expected {
			t.Errorf("For %d expected %s but got %s", tt.input, tt.expected, result)
		}
	}
}

func TestParseScanType(t *testing.T) {
	tests := []struct {
		input    string
		expected domain.ScanType
		hasError bool
	}{
		{"RequirementAnalysis", domain.ScanTypeRequirementAnalysis, false},
		{"ThreatModeling", domain.ScanTypeThreatModeling, false},
		{"SecuritySpecCheck", domain.ScanTypeSecuritySpecCheck, false},
		{"StaticCodeAnalysis", domain.ScanTypeStaticCodeAnalysis, false},
		{"ContainerImageScan", domain.ScanTypeContainerImageScan, false},
		{"HostSecurityCheck", domain.ScanTypeHostSecurityCheck, false},
		{"BlackBoxTesting", domain.ScanTypeBlackBoxTesting, false},
		{"PortScanning", domain.ScanTypePortScanning, false},
		{"DAST", domain.ScanTypeDast, false},
		{"SCA", domain.ScanTypeSca, false},
		{"SecretsDetection", domain.ScanTypeSecretsDetection, false},
		{"ComplianceAudit", domain.ScanTypeComplianceAudit, false},
		{"InvalidType", domain.ScanTypeUnknown, true},
	}

	for _, tt := range tests {
		result, err := domain.ParseScanType(tt.input)
		if (err != nil) != tt.hasError {
			t.Errorf("ParseScanType(%s) error = %v, wantErr %v", tt.input, err, tt.hasError)
			continue
		}
		if result != tt.expected {
			t.Errorf("ParseScanType(%s) = %v, want %v", tt.input, result, tt.expected)
		}
	}
}
