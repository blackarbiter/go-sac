package minio

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"go.uber.org/zap"
)

// ClientConfig 客户端配置
type ClientConfig struct {
	Endpoint       string
	AccessKey      string
	SecretKey      string
	UseSSL         bool
	Region         string
	RequestTimeout time.Duration
	DefaultBucket  string // 默认存储桶
}

// Client MinIO客户端
type Client struct {
	client *minio.Client
	logger *zap.Logger
	config ClientConfig
}

// NewClient 创建客户端实例
func NewClient(cfg ClientConfig, logger *zap.Logger) (*Client, error) {
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
		Region: cfg.Region,
	})
	if err != nil {
		return nil, fmt.Errorf("minio初始化失败: %w", err)
	}

	return &Client{
		client: client,
		logger: logger,
		config: cfg,
	}, nil
}

// PutObject 上传单个文件
func (c *Client) PutObject(ctx context.Context, objectPath string, reader io.Reader, size int64, opts minio.PutObjectOptions) (minio.UploadInfo, error) {
	bucket := c.getBucketFromPath(objectPath)
	objectName := c.getObjectNameFromPath(objectPath)

	// 确保存储桶存在
	if err := c.ensureBucketExists(ctx, bucket); err != nil {
		return minio.UploadInfo{}, err
	}

	return c.client.PutObject(ctx, bucket, objectName, reader, size, opts)
}

// GetObject 下载单个文件
func (c *Client) GetObject(ctx context.Context, objectPath string, opts minio.GetObjectOptions) (*minio.Object, error) {
	bucket := c.getBucketFromPath(objectPath)
	objectName := c.getObjectNameFromPath(objectPath)

	return c.client.GetObject(ctx, bucket, objectName, opts)
}

// PutObjects 批量上传文件
func (c *Client) PutObjects(ctx context.Context, objects map[string]io.Reader, opts minio.PutObjectOptions) error {
	for objectPath, reader := range objects {
		var size int64
		var err error

		// 尝试获取文件大小
		switch r := reader.(type) {
		case *os.File:
			fileInfo, err := r.Stat()
			if err != nil {
				return fmt.Errorf("获取文件大小失败: %w", err)
			}
			size = fileInfo.Size()
		case io.ReadSeeker:
			size, err = r.Seek(0, io.SeekEnd)
			if err != nil {
				return fmt.Errorf("获取文件大小失败: %w", err)
			}
			_, err = r.Seek(0, io.SeekStart)
			if err != nil {
				return fmt.Errorf("重置文件指针失败: %w", err)
			}
		default:
			return fmt.Errorf("reader必须实现io.ReadSeeker接口或为*os.File: %s", objectPath)
		}

		// 上传文件
		_, err = c.PutObject(ctx, objectPath, reader, size, opts)
		if err != nil {
			return fmt.Errorf("上传文件失败 %s: %w", objectPath, err)
		}
	}
	return nil
}

// GetObjects 批量下载文件
func (c *Client) GetObjects(ctx context.Context, objectPaths []string, opts minio.GetObjectOptions) (map[string]*minio.Object, error) {
	objects := make(map[string]*minio.Object)

	for _, objectPath := range objectPaths {
		obj, err := c.GetObject(ctx, objectPath, opts)
		if err != nil {
			// 关闭已打开的对象
			for _, o := range objects {
				o.Close()
			}
			return nil, fmt.Errorf("下载文件失败 %s: %w", objectPath, err)
		}
		objects[objectPath] = obj
	}

	return objects, nil
}

// PutLargeObject 分片上传大文件
func (c *Client) PutLargeObject(ctx context.Context, objectPath string, reader io.Reader, size int64, opts minio.PutObjectOptions) (minio.UploadInfo, error) {
	bucket := c.getBucketFromPath(objectPath)
	objectName := c.getObjectNameFromPath(objectPath)

	// 确保存储桶存在
	if err := c.ensureBucketExists(ctx, bucket); err != nil {
		return minio.UploadInfo{}, err
	}

	// 使用分片上传
	return c.client.PutObject(ctx, bucket, objectName, reader, size, opts)
}

// GetLargeObject 分片下载大文件
func (c *Client) GetLargeObject(ctx context.Context, objectPath string, writer io.Writer, opts minio.GetObjectOptions) error {
	obj, err := c.GetObject(ctx, objectPath, opts)
	if err != nil {
		return err
	}
	defer obj.Close()

	// 使用io.Copy进行分片下载
	_, err = io.Copy(writer, obj)
	return err
}

// ObjectResult 增强的返回结果类型，包含对象信息和目录前缀
type ObjectResult struct {
	Object minio.ObjectInfo
	IsDir  bool
}

// ListObjects 列出指定路径下的对象和目录
func (c *Client) ListObjects(ctx context.Context, path string, recursive bool) <-chan ObjectResult {
	resultChan := make(chan ObjectResult)

	go func() {
		defer close(resultChan)

		// 1. 增强路径解析
		bucket, prefix := parseMinioPath(path)
		if bucket == "" {
			bucket = c.config.DefaultBucket
		}

		// 2. 智能处理前缀格式
		processedPrefix := processPrefix(prefix, recursive)

		// 3. 配置列表参数
		opts := minio.ListObjectsOptions{
			Prefix:    processedPrefix,
			Recursive: recursive,
			UseV1:     false, // 强制使用v2 API
		}

		// 4. 执行列表操作
		objCh := c.client.ListObjects(ctx, bucket, opts)

		// 5. 处理返回结果
		for objInfo := range objCh {
			// 处理错误对象
			if objInfo.Err != nil {
				c.logger.Error("列出对象失败",
					zap.String("bucket", bucket),
					zap.String("prefix", processedPrefix),
					zap.Error(objInfo.Err))
				continue
			}

			// 处理普通文件对象
			if objInfo.Key != "" {
				// 检查是否为目录（通过检查Key是否以斜杠结尾）
				isDir := strings.HasSuffix(objInfo.Key, "/")
				resultChan <- ObjectResult{
					Object: objInfo,
					IsDir:  isDir,
				}
			}
		}
	}()

	return resultChan
}

// parseMinioPath 增强路径解析方法
func parseMinioPath(path string) (bucket, prefix string) {
	// 处理空路径的特殊情况
	if path == "" {
		return "", ""
	}

	// 分割bucket和前缀
	parts := strings.SplitN(path, "/", 2)
	if len(parts) == 1 {
		return parts[0], ""
	}

	// 处理根目录的情况
	if parts[1] == "" {
		return parts[0], ""
	}

	return parts[0], parts[1]
}

// processPrefix 智能处理前缀格式
func processPrefix(prefix string, recursive bool) string {
	// 空前缀直接返回
	if prefix == "" {
		return ""
	}

	// 递归模式不需要特殊处理
	if recursive {
		return prefix
	}

	// 非递归模式处理规则：
	switch {
	case strings.Contains(prefix, "."):
		// 包含扩展名的视为精确匹配
		return prefix
	case strings.HasSuffix(prefix, "/"):
		// 已有目录结尾符
		return prefix
	default:
		// 添加目录结尾符
		return prefix + "/"
	}
}

// getDelimiter 获取目录分隔符
func getDelimiter(recursive bool) string {
	if recursive {
		return "" // 递归模式不使用分隔符
	}
	return "/" // 非递归模式使用目录分隔符
}

// RemoveObject 删除单个对象
func (c *Client) RemoveObject(ctx context.Context, objectPath string, opts minio.RemoveObjectOptions) error {
	bucket := c.getBucketFromPath(objectPath)
	objectName := c.getObjectNameFromPath(objectPath)

	return c.client.RemoveObject(ctx, bucket, objectName, opts)
}

// RemoveObjects 批量删除对象
func (c *Client) RemoveObjects(ctx context.Context, objectPaths []string, opts minio.RemoveObjectOptions) error {
	objectsCh := make(chan minio.ObjectInfo)
	errorsCh := make(chan error, 1)

	go func() {
		defer close(objectsCh)
		for _, objectPath := range objectPaths {
			objectName := c.getObjectNameFromPath(objectPath)
			objectsCh <- minio.ObjectInfo{
				Key: objectName,
			}
		}
	}()

	// 启动删除操作
	removeErrorsCh := c.client.RemoveObjects(ctx, c.config.DefaultBucket, objectsCh, minio.RemoveObjectsOptions{})

	// 处理删除错误
	go func() {
		for err := range removeErrorsCh {
			errorsCh <- fmt.Errorf("删除对象失败: %w", err.Err)
		}
		close(errorsCh)
	}()

	// 等待第一个错误或完成
	return <-errorsCh
}

// StatObject 获取对象信息
func (c *Client) StatObject(ctx context.Context, objectPath string, opts minio.StatObjectOptions) (minio.ObjectInfo, error) {
	bucket := c.getBucketFromPath(objectPath)
	objectName := c.getObjectNameFromPath(objectPath)

	return c.client.StatObject(ctx, bucket, objectName, opts)
}

// CopyObject 复制对象
func (c *Client) CopyObject(ctx context.Context, srcPath, dstPath string, opts minio.CopyDestOptions, srcOpts minio.CopySrcOptions) (minio.UploadInfo, error) {
	srcBucket := c.getBucketFromPath(srcPath)
	srcObject := c.getObjectNameFromPath(srcPath)
	dstBucket := c.getBucketFromPath(dstPath)
	dstObject := c.getObjectNameFromPath(dstPath)

	// 设置源对象信息
	srcOpts.Bucket = srcBucket
	srcOpts.Object = srcObject

	// 设置目标对象信息
	opts.Bucket = dstBucket
	opts.Object = dstObject

	return c.client.CopyObject(ctx, opts, srcOpts)
}

// HealthCheck 健康检查
func (c *Client) HealthCheck(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, c.config.RequestTimeout)
	defer cancel()

	_, err := c.client.ListBuckets(ctx)
	if err != nil {
		c.logger.Error("存储服务不可达",
			zap.String("endpoint", c.config.Endpoint),
			zap.Error(err))
		return fmt.Errorf("存储服务健康检查失败: %w", err)
	}
	return nil
}

// 内部辅助方法

// getBucketFromPath 从路径中获取存储桶名称
func (c *Client) getBucketFromPath(objectPath string) string {
	parts := strings.SplitN(objectPath, "/", 2)
	if len(parts) > 1 {
		return parts[0]
	}
	return c.config.DefaultBucket
}

// getObjectNameFromPath 从路径中获取对象名称
func (c *Client) getObjectNameFromPath(objectPath string) string {
	parts := strings.SplitN(objectPath, "/", 2)
	if len(parts) > 1 {
		return parts[1]
	}
	return parts[0]
}

// ensureBucketExists 确保存储桶存在
func (c *Client) ensureBucketExists(ctx context.Context, bucket string) error {
	exists, err := c.client.BucketExists(ctx, bucket)
	if err != nil {
		return fmt.Errorf("检查存储桶是否存在失败: %w", err)
	}

	if !exists {
		err = c.client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{})
		if err != nil {
			return fmt.Errorf("创建存储桶失败: %w", err)
		}
	}

	return nil
}
