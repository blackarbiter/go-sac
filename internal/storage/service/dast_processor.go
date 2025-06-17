package service

import (
	"context"
	"encoding/json"

	"github.com/blackarbiter/go-sac/internal/storage/repository"
	"github.com/blackarbiter/go-sac/internal/storage/repository/model"
	"github.com/blackarbiter/go-sac/pkg/domain"
)

// DASTProcessor implements the Processor interface for DAST scan results
type DASTProcessor struct {
	repo repository.Repository
}

// NewDASTProcessor creates a new DASTProcessor instance
func NewDASTProcessor(repo repository.Repository) *DASTProcessor {
	return &DASTProcessor{repo: repo}
}

// Process handles DAST scan results
func (p *DASTProcessor) Process(ctx context.Context, result *domain.ScanResult) error {
	dastResult := &model.DASTModel{
		TaskID:    result.TaskID,
		AssetID:   result.AssetID,
		AssetType: string(result.AssetType),
		Status:    result.Status,
		Error:     result.Error,
	}

	// Parse DAST specific fields from result.Result
	if result.Status == "success" {
		jsonBytes, err := json.Marshal(result.Result)
		if err != nil {
			return err
		}
		if err := json.Unmarshal(jsonBytes, dastResult); err != nil {
			return err
		}
	}

	return p.repo.CreateDAST(ctx, dastResult)
}

// GetScanType returns the scan type this processor handles
func (p *DASTProcessor) GetScanType() domain.ScanType {
	return domain.ScanTypeDast
}

// Query retrieves DAST scan results by task ID
func (p *DASTProcessor) Query(ctx context.Context, taskID string) (interface{}, error) {
	return p.repo.FindDASTByTaskID(ctx, taskID)
}

// BatchQuery retrieves DAST scan results by multiple task IDs
func (p *DASTProcessor) BatchQuery(ctx context.Context, taskIDs []string) ([]interface{}, error) {
	results, err := p.repo.FindDASTByTaskIDs(ctx, taskIDs)
	if err != nil {
		return nil, err
	}

	// Convert []*model.DASTModel to []interface{}
	interfaceResults := make([]interface{}, len(results))
	for i, result := range results {
		interfaceResults[i] = result
	}
	return interfaceResults, nil
}
