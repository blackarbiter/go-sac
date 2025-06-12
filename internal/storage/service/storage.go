package service

import (
	"context"
	"errors"
	"fmt"
	"mime/multipart"
	"path/filepath"
	"time"

	"github.com/blackarbiter/go-sac/internal/storage/dto"
	"github.com/blackarbiter/go-sac/internal/storage/repository"
	"github.com/blackarbiter/go-sac/pkg/storage/minio"
)

// StorageService 定义存储服务接口
type StorageService interface {
	CreateStorage(ctx context.Context, req *dto.StorageRequest) (*dto.StorageResponse, error)
	GetStorage(ctx context.Context, id string) (*dto.StorageResponse, error)
	UpdateStorage(ctx context.Context, id string, req *dto.StorageRequest) (*dto.StorageResponse, error)
	DeleteStorage(ctx context.Context, id string) error
	ListStorages(ctx context.Context, params *dto.StorageQueryParams) (*dto.StorageListResponse, error)
	BatchCreateStorages(ctx context.Context, req *dto.BatchCreateStorageRequest) ([]string, error)
	UploadFile(ctx context.Context, id string, file *multipart.FileHeader) error
	DownloadFile(ctx context.Context, id string) (string, error)
}

// storageService 是StorageService的具体实现
type storageService struct {
	repo    repository.StorageRepository
	storage *minio.Storage
}

// NewStorageService 创建一个新的存储服务实例
func NewStorageService(repo repository.StorageRepository, storage *minio.Storage) StorageService {
	return &storageService{
		repo:    repo,
		storage: storage,
	}
}

// convertToDTO 将存储实体转换为DTO
func convertToDTO(storage *repository.Storage) *dto.StorageResponse {
	return &dto.StorageResponse{
		ID:        storage.ID,
		Path:      storage.Path,
		Type:      storage.Type,
		Status:    storage.Status,
		Metadata:  storage.Metadata,
		CreatedAt: storage.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: storage.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

// CreateStorage 创建存储
func (s *storageService) CreateStorage(ctx context.Context, req *dto.StorageRequest) (*dto.StorageResponse, error) {
	storage := &repository.Storage{
		Path:     req.Path,
		Type:     req.Type,
		Status:   "pending",
		Metadata: req.Metadata,
	}

	if err := s.repo.Create(ctx, storage); err != nil {
		return nil, fmt.Errorf("failed to create storage: %w", err)
	}

	return convertToDTO(storage), nil
}

// GetStorage 获取存储信息
func (s *storageService) GetStorage(ctx context.Context, id string) (*dto.StorageResponse, error) {
	storage, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrStorageNotFound) {
			return nil, fmt.Errorf("storage not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get storage: %w", err)
	}

	return convertToDTO(storage), nil
}

// UpdateStorage 更新存储信息
func (s *storageService) UpdateStorage(ctx context.Context, id string, req *dto.StorageRequest) (*dto.StorageResponse, error) {
	storage, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrStorageNotFound) {
			return nil, fmt.Errorf("storage not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get storage: %w", err)
	}

	storage.Path = req.Path
	storage.Type = req.Type
	storage.Metadata = req.Metadata

	if err := s.repo.Update(ctx, storage); err != nil {
		return nil, fmt.Errorf("failed to update storage: %w", err)
	}

	return convertToDTO(storage), nil
}

// DeleteStorage 删除存储
func (s *storageService) DeleteStorage(ctx context.Context, id string) error {
	storage, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrStorageNotFound) {
			return fmt.Errorf("storage not found: %w", err)
		}
		return fmt.Errorf("failed to get storage: %w", err)
	}

	// 如果存储有文件，先删除文件
	if storage.Status == "completed" {
		if err := s.storage.DeleteObject(ctx, storage.Path); err != nil {
			return fmt.Errorf("failed to delete file: %w", err)
		}
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete storage: %w", err)
	}

	return nil
}

// ListStorages 列出存储
func (s *storageService) ListStorages(ctx context.Context, params *dto.StorageQueryParams) (*dto.StorageListResponse, error) {
	var storages []*repository.Storage
	var total int64
	var err error

	offset := (params.Page - 1) * params.Size

	switch {
	case params.Status != "":
		storages, total, err = s.repo.FindByStatus(ctx, params.Status, params.Size, offset)
	case params.Type != "":
		storages, total, err = s.repo.FindByType(ctx, params.Type, params.Size, offset)
	default:
		storages, total, err = s.repo.List(ctx, params.Size, offset)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to list storages: %w", err)
	}

	responses := make([]dto.StorageResponse, len(storages))
	for i, storage := range storages {
		responses[i] = *convertToDTO(storage)
	}

	return &dto.StorageListResponse{
		Total: total,
		Items: responses,
	}, nil
}

// BatchCreateStorages 批量创建存储
func (s *storageService) BatchCreateStorages(ctx context.Context, req *dto.BatchCreateStorageRequest) ([]string, error) {
	storages := make([]*repository.Storage, len(req.Storages))
	for i, storageReq := range req.Storages {
		storages[i] = &repository.Storage{
			Path:     storageReq.Path,
			Type:     storageReq.Type,
			Status:   "pending",
			Metadata: storageReq.Metadata,
		}
	}

	if err := s.repo.BatchCreate(ctx, storages); err != nil {
		return nil, fmt.Errorf("failed to batch create storages: %w", err)
	}

	ids := make([]string, len(storages))
	for i, storage := range storages {
		ids[i] = storage.ID
	}

	return ids, nil
}

// UploadFile 上传文件
func (s *storageService) UploadFile(ctx context.Context, id string, file *multipart.FileHeader) error {
	storage, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrStorageNotFound) {
			return fmt.Errorf("storage not found: %w", err)
		}
		return fmt.Errorf("failed to get storage: %w", err)
	}

	// 生成文件路径
	ext := filepath.Ext(file.Filename)
	path := fmt.Sprintf("%s/%s%s", storage.Type, storage.ID, ext)

	// 上传文件到存储服务
	if err := s.storage.UploadFile(ctx, path, file); err != nil {
		return fmt.Errorf("failed to upload file: %w", err)
	}

	// 更新存储记录
	storage.Path = path
	storage.Status = "completed"
	storage.Metadata["filename"] = file.Filename
	storage.Metadata["size"] = fmt.Sprintf("%d", file.Size)
	storage.Metadata["content_type"] = file.Header.Get("Content-Type")

	if err := s.repo.Update(ctx, storage); err != nil {
		return fmt.Errorf("failed to update storage: %w", err)
	}

	return nil
}

// DownloadFile 下载文件
func (s *storageService) DownloadFile(ctx context.Context, id string) (string, error) {
	storage, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrStorageNotFound) {
			return "", fmt.Errorf("storage not found: %w", err)
		}
		return "", fmt.Errorf("failed to get storage: %w", err)
	}

	if storage.Status != "completed" {
		return "", fmt.Errorf("storage is not completed")
	}

	// 生成临时下载URL
	url, err := s.storage.GetPresignedURL(ctx, storage.Path, 24*time.Hour)
	if err != nil {
		return "", fmt.Errorf("failed to get presigned URL: %w", err)
	}

	return url, nil
}
