package repository

import (
	"context"

	"github.com/blackarbiter/go-sac/internal/storage/repository/model"
)

// Repository defines the interface for storage operations
type Repository interface {
	// AutoMigrate performs database migrations
	AutoMigrate() error

	// SAST operations
	CreateSAST(ctx context.Context, result *model.SASTModel) error
	BatchCreateSAST(ctx context.Context, results []*model.SASTModel) error
	FindSASTByTaskID(ctx context.Context, taskID string) (*model.SASTModel, error)
	FindSASTByTaskIDs(ctx context.Context, taskIDs []string) ([]*model.SASTModel, error)
	UpdateSAST(ctx context.Context, result *model.SASTModel) error

	// DAST operations
	CreateDAST(ctx context.Context, result *model.DASTModel) error
	BatchCreateDAST(ctx context.Context, results []*model.DASTModel) error
	FindDASTByTaskID(ctx context.Context, taskID string) (*model.DASTModel, error)
	FindDASTByTaskIDs(ctx context.Context, taskIDs []string) ([]*model.DASTModel, error)
	UpdateDAST(ctx context.Context, result *model.DASTModel) error

	// SCA operations
	CreateSCA(ctx context.Context, result *model.SCAModel) error
	BatchCreateSCA(ctx context.Context, results []*model.SCAModel) error
	FindSCAByTaskID(ctx context.Context, taskID string) (*model.SCAModel, error)
	FindSCAByTaskIDs(ctx context.Context, taskIDs []string) ([]*model.SCAModel, error)
	UpdateSCA(ctx context.Context, result *model.SCAModel) error
}
