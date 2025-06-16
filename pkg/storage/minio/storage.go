package minio

import (
	"context"
	"mime/multipart"
	"net/url"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// Storage 表示 MinIO 存储服务
type Storage struct {
	client     *minio.Client
	bucketName string
}

// NewStorage 创建一个新的 MinIO 存储服务实例
func NewStorage(endpoint, accessKey, secretKey, bucketName string, useSSL bool) (*Storage, error) {
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, err
	}

	return &Storage{
		client:     client,
		bucketName: bucketName,
	}, nil
}

// UploadFile 上传文件
func (s *Storage) UploadFile(ctx context.Context, path string, file *multipart.FileHeader) error {
	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	_, err = s.client.PutObject(ctx, s.bucketName, path, src, file.Size, minio.PutObjectOptions{
		ContentType: file.Header.Get("Content-Type"),
	})
	return err
}

// DeleteObject 删除对象
func (s *Storage) DeleteObject(ctx context.Context, path string) error {
	return s.client.RemoveObject(ctx, s.bucketName, path, minio.RemoveObjectOptions{})
}

// GetPresignedURL 获取预签名URL
func (s *Storage) GetPresignedURL(ctx context.Context, path string, expiry time.Duration) (string, error) {
	reqParams := make(url.Values)
	url, err := s.client.PresignedGetObject(ctx, s.bucketName, path, expiry, reqParams)
	if err != nil {
		return "", err
	}
	return url.String(), nil
}
