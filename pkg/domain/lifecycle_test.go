package domain_test

import (
	"github.com/blackarbiter/go-sac/pkg/domain"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLifecyclePhaseString(t *testing.T) {
	tests := []struct {
		phase    domain.LifecyclePhase
		expected string
	}{
		{domain.LifecyclePhasePlanning, "Planning"},
		{domain.LifecyclePhaseDevelopment, "Development"},
		{domain.LifecyclePhaseBuilding, "Building"},
		{domain.LifecyclePhaseTesting, "Testing"},
		{domain.LifecyclePhaseRelease, "Release"},
		{domain.LifecyclePhaseDelivery, "Delivery"},
		{domain.LifecyclePhaseDeployment, "Deployment"},
		{domain.LifecyclePhaseOperation, "Operation"},
		{domain.LifecyclePhaseMonitoring, "Monitoring"},
		{domain.LifecyclePhaseFeedback, "Feedback"},
		{domain.LifecyclePhaseSecurityHardening, "SecurityHardening"},
		{domain.LifecyclePhaseComplianceAudit, "ComplianceAudit"},
		{100, "Unknown"},
	}

	for _, tt := range tests {
		assert.Equal(t, tt.expected, tt.phase.String())
	}
}

func TestParseLifecyclePhase(t *testing.T) {
	validCases := map[string]domain.LifecyclePhase{
		"Planning":          domain.LifecyclePhasePlanning,
		"SecurityHardening": domain.LifecyclePhaseSecurityHardening,
		"ComplianceAudit":   domain.LifecyclePhaseComplianceAudit,
	}

	for input, expected := range validCases {
		result, err := domain.ParseLifecyclePhase(input)
		assert.NoError(t, err)
		assert.Equal(t, expected, result)
	}

	invalidCases := []string{"", "InvalidPhase", "planning"}
	for _, input := range invalidCases {
		_, err := domain.ParseLifecyclePhase(input)
		assert.Error(t, err)
	}
}
