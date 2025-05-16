package minio

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestPresignedURL(t *testing.T) {
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

	// 测试生成上传URL
	t.Run("GetPresignedPutURL", func(t *testing.T) {
		// 生成PUT URL
		putURL, err := client.GetPresignedPutURL(
			ctx,
			"test-bucket",
			"test-object.txt",
			PresignConfig{
				Expires: time.Hour,
			},
		)
		assert.NoError(t, err)
		assert.NotNil(t, putURL)
		assert.Contains(t, putURL.Query().Get("x-security-sig"), "")

		// 使用URL上传文件
		content := []byte("test content")
		req, err := http.NewRequest(http.MethodPut, putURL.String(), bytes.NewReader(content))
		assert.NoError(t, err)
		req.Header.Set("Content-Type", "text/plain")

		// 验证签名
		assert.True(t, client.VerifyPresignedRequest(req))

		// 发送请求
		resp, err := http.DefaultClient.Do(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		resp.Body.Close()

		// 验证文件已上传
		obj, err := client.GetObject(ctx, "test-bucket/test-object.txt", minio.GetObjectOptions{})
		assert.NoError(t, err)
		defer obj.Close()

		downloaded, err := io.ReadAll(obj)
		assert.NoError(t, err)
		assert.Equal(t, content, downloaded)
	})

	// 测试生成下载URL
	t.Run("GetPresignedGetURL", func(t *testing.T) {
		// 先上传一个测试文件
		content := []byte("test content for download")
		_, err := client.PutObject(
			ctx,
			"test-bucket/download-test.txt",
			bytes.NewReader(content),
			int64(len(content)),
			minio.PutObjectOptions{},
		)
		assert.NoError(t, err)

		// 生成GET URL
		getURL, err := client.GetPresignedGetURL(
			ctx,
			"test-bucket",
			"download-test.txt",
			PresignConfig{
				Expires: time.Hour,
			},
		)
		assert.NoError(t, err)
		assert.NotNil(t, getURL)
		assert.Contains(t, getURL.Query().Get("x-security-sig"), "")

		// 使用URL下载文件
		req, err := http.NewRequest(http.MethodGet, getURL.String(), nil)
		assert.NoError(t, err)

		// 验证签名
		assert.True(t, client.VerifyPresignedRequest(req))

		// 发送请求
		resp, err := http.DefaultClient.Do(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		// 验证下载的内容
		downloaded, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		assert.NoError(t, err)
		assert.Equal(t, content, downloaded)
	})

	// 测试URL安全性
	t.Run("URLSecurity", func(t *testing.T) {
		// 生成PUT URL
		putURL, err := client.GetPresignedPutURL(
			ctx,
			"test-bucket",
			"security-test.txt",
			PresignConfig{
				Expires: time.Hour,
			},
		)
		assert.NoError(t, err)

		// 测试篡改签名
		req := &http.Request{URL: putURL}
		assert.True(t, client.VerifyPresignedRequest(req))

		// 篡改签名
		q := putURL.Query()
		q.Set("x-security-sig", "invalid")
		req.URL.RawQuery = q.Encode()
		assert.False(t, client.VerifyPresignedRequest(req))

		// 测试篡改路径
		req.URL.Path = "/test-bucket/another-object.txt"
		assert.False(t, client.VerifyPresignedRequest(req))

		// 测试篡改查询参数
		q = req.URL.Query()
		q.Set("x-amz-date", "invalid-date")
		req.URL.RawQuery = q.Encode()
		assert.False(t, client.VerifyPresignedRequest(req))
	})

	// 测试URL过期
	t.Run("URLExpiration", func(t *testing.T) {
		// 生成一个1秒后过期的URL
		putURL, err := client.GetPresignedPutURL(
			ctx,
			"test-bucket",
			"expiration-test.txt",
			PresignConfig{
				Expires: time.Second,
			},
		)
		assert.NoError(t, err)

		// 等待URL过期
		time.Sleep(time.Second * 2)

		// 尝试使用过期的URL
		content := []byte("test content")
		req, err := http.NewRequest(http.MethodPut, putURL.String(), bytes.NewReader(content))
		assert.NoError(t, err)
		req.Header.Set("Content-Type", "text/plain")

		resp, err := http.DefaultClient.Do(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusForbidden, resp.StatusCode)
		resp.Body.Close()
	})

	// 测试自定义请求头
	t.Run("CustomHeaders", func(t *testing.T) {
		headers := http.Header{}
		headers.Set("Content-Type", "application/json")
		headers.Set("x-amz-meta-custom", "test-value")

		putURL, err := client.GetPresignedPutURL(
			ctx,
			"test-bucket",
			"headers-test.txt",
			PresignConfig{
				Expires: time.Hour,
				Headers: headers,
			},
		)
		assert.NoError(t, err)

		// 使用URL上传文件
		content := []byte(`{"test": "content"}`)
		req, err := http.NewRequest(http.MethodPut, putURL.String(), bytes.NewReader(content))
		assert.NoError(t, err)
		req.Header = headers

		// 验证签名
		assert.True(t, client.VerifyPresignedRequest(req))

		// 发送请求
		resp, err := http.DefaultClient.Do(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		resp.Body.Close()

		// 验证元数据
		info, err := client.StatObject(ctx, "test-bucket/headers-test.txt", minio.StatObjectOptions{})
		assert.NoError(t, err)
		assert.Equal(t, "application/json", info.ContentType)
		assert.Equal(t, "test-value", info.UserMetadata["x-amz-meta-custom"])
	})
}

// 添加一个独立的单元测试，不依赖MinIO服务器
func TestSignWithHmac(t *testing.T) {
	// 创建客户端
	cfg := ClientConfig{
		Endpoint:  "localhost:9000",
		AccessKey: "admin",
		SecretKey: "test-secret-key",
	}
	client, _ := NewClient(cfg, zap.NewNop())

	// 创建测试URL
	testURL, _ := url.Parse("https://test-minio.example.com/test-bucket/test-object?X-Amz-Algorithm=AWS4-HMAC-SHA256")

	// 签名URL
	signedURL := client.signWithHmac(testURL, PresignConfig{
		Expires: time.Hour,
	})

	// 验证签名存在
	assert.NotEmpty(t, signedURL.Query().Get("x-security-sig"))

	// 验证签名
	req := &http.Request{URL: signedURL}
	assert.True(t, client.VerifyPresignedRequest(req))

	// 篡改签名应失败
	q := signedURL.Query()
	q.Set("x-security-sig", "invalid")
	signedURL.RawQuery = q.Encode()
	req = &http.Request{URL: signedURL}
	assert.False(t, client.VerifyPresignedRequest(req))
}
