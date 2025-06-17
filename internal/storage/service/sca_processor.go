package service

import (
	"context"
	"encoding/json"

	"github.com/blackarbiter/go-sac/internal/storage/repository"
	"github.com/blackarbiter/go-sac/internal/storage/repository/model"
	"github.com/blackarbiter/go-sac/pkg/domain"
)

// SCAProcessor implements the Processor interface for SCA scan results
type SCAProcessor struct {
	repo repository.Repository
}

// NewSCAProcessor creates a new SCAProcessor instance
func NewSCAProcessor(repo repository.Repository) *SCAProcessor {
	return &SCAProcessor{repo: repo}
}

// Process handles SCA scan results
func (p *SCAProcessor) Process(ctx context.Context, result *domain.ScanResult) error {
	scaResult := &model.SCAModel{
		TaskID:    result.TaskID,
		AssetID:   result.AssetID,
		AssetType: string(result.AssetType),
		Status:    result.Status,
		Error:     result.Error,
	}

	// Parse SCA specific fields from result.Result
	if result.Status == "success" {
		jsonBytes, err := json.Marshal(result.Result)
		if err != nil {
			return err
		}
		if err := json.Unmarshal(jsonBytes, scaResult); err != nil {
			return err
		}
	}

	return p.repo.CreateSCA(ctx, scaResult)
}

// GetScanType returns the scan type this processor handles
func (p *SCAProcessor) GetScanType() domain.ScanType {
	return domain.ScanTypeSca
}

// Query retrieves SCA scan results by task ID
func (p *SCAProcessor) Query(ctx context.Context, taskID string) (interface{}, error) {
	return p.repo.FindSCAByTaskID(ctx, taskID)
}

// BatchQuery retrieves SCA scan results by multiple task IDs
func (p *SCAProcessor) BatchQuery(ctx context.Context, taskIDs []string) ([]interface{}, error) {
	results, err := p.repo.FindSCAByTaskIDs(ctx, taskIDs)
	if err != nil {
		return nil, err
	}

	// Convert []*model.SCAModel to []interface{}
	interfaceResults := make([]interface{}, len(results))
	for i, result := range results {
		interfaceResults[i] = result
	}
	return interfaceResults, nil
}
