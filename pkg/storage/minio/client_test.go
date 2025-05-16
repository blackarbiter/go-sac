package minio

import (
	"bytes"
	"context"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestClient(t *testing.T) {
	// 跳过真实服务器测试
	t.Skip("此测试需要真实的MinIO服务器")

	cfg := ClientConfig{
		Endpoint:       "localhost:9000",
		AccessKey:      "admin",
		SecretKey:      "1234qwer",
		DefaultBucket:  "test-bucket",
		RequestTimeout: time.Second * 5,
	}

	client, err := NewClient(cfg, zap.NewNop())
	assert.NoError(t, err)

	ctx := context.Background()

	// 测试上传单个文件
	t.Run("PutObject", func(t *testing.T) {
		content := []byte("test content")
		reader := bytes.NewReader(content)

		info, err := client.PutObject(ctx, "test-bucket/test-object.txt", reader, int64(len(content)), minio.PutObjectOptions{})
		assert.NoError(t, err)
		assert.NotEmpty(t, info.ETag)
	})

	// 测试下载单个文件
	t.Run("GetObject", func(t *testing.T) {
		obj, err := client.GetObject(ctx, "test-bucket/test-object.txt", minio.GetObjectOptions{})
		assert.NoError(t, err)
		defer obj.Close()

		content, err := io.ReadAll(obj)
		assert.NoError(t, err)
		assert.Equal(t, "test content", string(content))
	})

	// 测试批量上传文件
	t.Run("PutObjects", func(t *testing.T) {
		objects := map[string]io.Reader{
			"test-bucket/test1.txt": strings.NewReader("test1"),
			"test-bucket/test2.txt": strings.NewReader("test2"),
		}

		err := client.PutObjects(ctx, objects, minio.PutObjectOptions{})
		assert.NoError(t, err)
	})

	// 测试批量下载文件
	t.Run("GetObjects", func(t *testing.T) {
		objectPaths := []string{
			"test-bucket/test1.txt",
			"test-bucket/test2.txt",
		}

		objects, err := client.GetObjects(ctx, objectPaths, minio.GetObjectOptions{})
		assert.NoError(t, err)
		assert.Len(t, objects, 2)

		// 读取并验证内容
		for _, obj := range objects {
			defer obj.Close()
			content, err := io.ReadAll(obj)
			assert.NoError(t, err)
			assert.Contains(t, []string{"test1", "test2"}, string(content))
		}
	})

	// 测试大文件上传
	t.Run("PutLargeObject", func(t *testing.T) {
		// 创建1MB的测试数据
		content := bytes.Repeat([]byte("test"), 256*1024)
		reader := bytes.NewReader(content)

		info, err := client.PutLargeObject(ctx, "test-bucket/large-file.txt", reader, int64(len(content)), minio.PutObjectOptions{})
		assert.NoError(t, err)
		assert.NotEmpty(t, info.ETag)
	})

	// 测试大文件下载
	t.Run("GetLargeObject", func(t *testing.T) {
		var buf bytes.Buffer
		err := client.GetLargeObject(ctx, "test-bucket/large-file.txt", &buf, minio.GetObjectOptions{})
		assert.NoError(t, err)
		assert.Equal(t, 1024*1024, buf.Len())
	})

	// 测试列出对象
	t.Run("ListObjects", func(t *testing.T) {
		objects := client.ListObjects(ctx, "test-bucket/", true)
		count := 0
		for obj := range objects {
			assert.NotEmpty(t, obj.Key)
			count++
		}
		assert.Greater(t, count, 0)
	})

	// 测试删除对象
	t.Run("RemoveObject", func(t *testing.T) {
		err := client.RemoveObject(ctx, "test-bucket/test-object.txt", minio.RemoveObjectOptions{})
		assert.NoError(t, err)
	})

	// 测试批量删除对象
	t.Run("RemoveObjects", func(t *testing.T) {
		objectPaths := []string{
			"test-bucket/test1.txt",
			"test-bucket/test2.txt",
		}

		err := client.RemoveObjects(ctx, objectPaths, minio.RemoveObjectOptions{})
		assert.NoError(t, err)
	})

	// 测试获取对象信息
	t.Run("StatObject", func(t *testing.T) {
		info, err := client.StatObject(ctx, "test-bucket/large-file.txt", minio.StatObjectOptions{})
		assert.NoError(t, err)
		assert.Equal(t, int64(1024*1024), info.Size)
	})

	// 测试复制对象
	t.Run("CopyObject", func(t *testing.T) {
		info, err := client.CopyObject(ctx,
			"test-bucket/large-file.txt",
			"test-bucket/large-file-copy.txt",
			minio.CopyDestOptions{},
			minio.CopySrcOptions{},
		)
		assert.NoError(t, err)
		assert.NotEmpty(t, info.ETag)
	})

	// 测试健康检查
	t.Run("HealthCheck", func(t *testing.T) {
		err := client.HealthCheck(ctx)
		assert.NoError(t, err)
	})
}
