package service

import (
	"context"

	"github.com/blackarbiter/go-sac/pkg/domain"
)

// StorageProcessor defines the interface for scan result processing
type StorageProcessor interface {
	// Process handles the scan result
	Process(ctx context.Context, result *domain.ScanResult) error
	// GetScanType returns the scan type this processor handles
	GetScanType() domain.ScanType
	// Query retrieves scan results by task ID
	Query(ctx context.Context, taskID string) (interface{}, error)
	// BatchQuery retrieves scan results by multiple task IDs
	BatchQuery(ctx context.Context, taskIDs []string) ([]interface{}, error)
}

// StorageProcessorFactory defines the interface for processor factory
type StorageProcessorFactory interface {
	// GetProcessor returns the processor for the given scan type
	GetProcessor(scanType domain.ScanType) (StorageProcessor, error)
	// RegisterProcessor registers a processor for a scan type
	RegisterProcessor(scanType domain.ScanType, processor StorageProcessor)
}
