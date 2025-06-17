package service

import (
	"context"
	"encoding/json"

	"github.com/blackarbiter/go-sac/internal/storage/repository"
	"github.com/blackarbiter/go-sac/internal/storage/repository/model"
	"github.com/blackarbiter/go-sac/pkg/domain"
)

// SASTProcessor implements the Processor interface for SAST scan results
type SASTProcessor struct {
	repo repository.Repository
}

// NewSASTProcessor creates a new SASTProcessor instance
func NewSASTProcessor(repo repository.Repository) *SASTProcessor {
	return &SASTProcessor{repo: repo}
}

// Process handles SAST scan results
func (p *SASTProcessor) Process(ctx context.Context, result *domain.ScanResult) error {
	sastResult := &model.SASTModel{
		TaskID:    result.TaskID,
		AssetID:   result.AssetID,
		AssetType: result.AssetType.String(),
		Status:    result.Status,
		Error:     result.Error,
	}

	// Parse SAST specific fields from result.Result
	if result.Status == "success" {
		jsonBytes, err := json.Marshal(result.Result)
		if err != nil {
			return err
		}
		if err := json.Unmarshal(jsonBytes, sastResult); err != nil {
			return err
		}
	}

	return p.repo.CreateSAST(ctx, sastResult)
}

// GetScanType returns the scan type this processor handles
func (p *SASTProcessor) GetScanType() domain.ScanType {
	return domain.ScanTypeStaticCodeAnalysis
}

// Query retrieves SAST scan results by task ID
func (p *SASTProcessor) Query(ctx context.Context, taskID string) (interface{}, error) {
	return p.repo.FindSASTByTaskID(ctx, taskID)
}

// BatchQuery retrieves SAST scan results by multiple task IDs
func (p *SASTProcessor) BatchQuery(ctx context.Context, taskIDs []string) ([]interface{}, error) {
	results, err := p.repo.FindSASTByTaskIDs(ctx, taskIDs)
	if err != nil {
		return nil, err
	}

	// Convert []*model.SASTModel to []interface{}
	interfaceResults := make([]interface{}, len(results))
	for i, result := range results {
		interfaceResults[i] = result
	}
	return interfaceResults, nil
}
