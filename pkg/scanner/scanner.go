package scanner

import (
	"context"

	"github.com/blackarbiter/go-sac/pkg/domain"
)

// Scanner defines the interface for all scanner implementations
type Scanner interface {
	// Scan performs the actual scanning operation
	Scan(ctx context.Context, task *domain.ScanTaskPayload) (*domain.ScanResult, error)
}
