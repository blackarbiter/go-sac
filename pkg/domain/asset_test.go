package domain_test

import (
	"github.com/blackarbiter/go-sac/pkg/domain"
	"testing"
)

func TestAssetTypeString(t *testing.T) {
	tests := []struct {
		input    domain.AssetType
		expected string
	}{
		{domain.AssetTypeRequirement, "Requirement"},
		{domain.AssetTypeDesignDocument, "DesignDocument"},
		{domain.AssetTypeRepository, "Repository"},
		{domain.AssetTypeUploadedFile, "UploadedFile"},
		{domain.AssetTypeImage, "Image"},
		{domain.AssetTypeDomain, "Domain"},
		{domain.AssetTypeIP, "IP"},
		{domain.AssetTypeUnknown, "Unknown"},
		{100, "Unknown"}, // 测试无效值
	}

	for _, tt := range tests {
		result := tt.input.String()
		if result != tt.expected {
			t.Errorf("For %d expected %s but got %s", tt.input, tt.expected, result)
		}
	}
}

func TestParseAssetType(t *testing.T) {
	tests := []struct {
		input    string
		expected domain.AssetType
		hasError bool
	}{
		{"Requirement", domain.AssetTypeRequirement, false},
		{"DesignDocument", domain.AssetTypeDesignDocument, false},
		{"Repository", domain.AssetTypeRepository, false},
		{"UploadedFile", domain.AssetTypeUploadedFile, false},
		{"Image", domain.AssetTypeImage, false},
		{"Domain", domain.AssetTypeDomain, false},
		{"IP", domain.AssetTypeIP, false},
		{"InvalidType", domain.AssetTypeUnknown, true},
		{"", domain.AssetTypeUnknown, true},
	}

	for _, tt := range tests {
		result, err := domain.ParseAssetType(tt.input)
		if (err != nil) != tt.hasError {
			t.Errorf("ParseAssetType(%s) error = %v, wantErr %v", tt.input, err, tt.hasError)
			continue
		}
		if result != tt.expected {
			t.Errorf("ParseAssetType(%s) = %v, want %v", tt.input, result, tt.expected)
		}
	}
}
