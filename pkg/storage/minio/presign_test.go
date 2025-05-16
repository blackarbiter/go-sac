package minio

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestPresignedURL(t *testing.T) {
	// 使用本地MinIO服务器配置
	cfg := ClientConfig{
		Endpoint:       "localhost:9000",
		AccessKey:      "admin",
		SecretKey:      "1234qwer",
		DefaultBucket:  "test-bucket",
		RequestTimeout: time.Second * 5,
	}

	// 创建MinIO客户端
	client, err := NewClient(cfg, zap.NewNop())
	assert.NoError(t, err)

	ctx := context.Background()

	// 确保测试存储桶存在
	err = client.ensureBucketExists(ctx, "test-bucket")
	assert.NoError(t, err)

	// 创建临时测试目录
	tempDir := t.TempDir()
	t.Logf("测试目录: %s", tempDir)

	// 1. 创建测试文件
	uploadDir := filepath.Join(tempDir, "upload")
	err = os.MkdirAll(uploadDir, 0755)
	assert.NoError(t, err)

	uploadTestFile := filepath.Join(uploadDir, "test-upload.txt")
	uploadContent := []byte("这是一个通过预签名URL上传的测试文件内容")
	err = os.WriteFile(uploadTestFile, uploadContent, 0644)
	assert.NoError(t, err)
	t.Logf("创建上传测试文件: %s (大小: %d 字节)", uploadTestFile, len(uploadContent))

	// 2. 创建下载目录
	downloadDir := filepath.Join(tempDir, "download")
	err = os.MkdirAll(downloadDir, 0755)
	assert.NoError(t, err)
	t.Logf("创建下载目录: %s", downloadDir)

	// 测试预签名上传与下载
	t.Run("预签名上传与下载", func(t *testing.T) {
		// 定义文件对象信息
		bucket := "test-bucket"
		objectKey := fmt.Sprintf("presigned-test-%d.txt", time.Now().Unix())
		t.Logf("测试对象: %s/%s", bucket, objectKey)

		// 步骤1: 生成上传URL
		t.Log("步骤1: 生成预签名上传URL")
		putURL, err := client.GetPresignedPutURL(ctx, bucket, objectKey, PresignConfig{
			Expires:     time.Hour,
			ContentType: "text/plain",
			Metadata: map[string]string{
				"description": "测试文件",
				"created-by":  "presign-test",
			},
		})
		assert.NoError(t, err)
		t.Logf("生成上传URL: %s", putURL.String())

		// 步骤2: 使用预签名URL上传文件
		t.Log("步骤2: 使用预签名URL上传文件")

		// 读取文件内容到内存
		fileContent, err := os.ReadFile(uploadTestFile)
		assert.NoError(t, err)

		// 创建HTTP请求，使用内存中的内容而不是文件流
		req, err := http.NewRequest(http.MethodPut, putURL.String(), bytes.NewReader(fileContent))
		assert.NoError(t, err)
		req.Header.Set("Content-Type", "text/plain")
		req.Header.Set("Content-Length", fmt.Sprintf("%d", len(fileContent)))

		// 添加元数据头
		req.Header.Set("x-amz-meta-description", "测试文件")
		req.Header.Set("x-amz-meta-created-by", "presign-test")

		t.Logf("设置Content-Length: %d", len(fileContent))

		// 发送请求
		resp, err := http.DefaultClient.Do(req)
		assert.NoError(t, err)
		defer resp.Body.Close()

		// 验证上传成功
		assert.Equal(t, http.StatusOK, resp.StatusCode, "上传响应状态码应为200")
		t.Logf("上传成功，状态码: %d", resp.StatusCode)

		// 步骤3: 验证文件已上传到MinIO
		t.Log("步骤3: 验证文件已上传到MinIO")
		objectInfo, err := client.StatObject(ctx, bucket+"/"+objectKey, minio.StatObjectOptions{})
		assert.NoError(t, err)
		assert.Equal(t, int64(len(uploadContent)), objectInfo.Size, "上传文件大小不匹配")
		assert.Equal(t, "text/plain", objectInfo.ContentType, "文件内容类型不匹配")
		assert.Equal(t, "测试文件", objectInfo.UserMetadata["Description"], "文件元数据不匹配")
		t.Logf("文件成功上传到MinIO，大小: %d 字节, 内容类型: %s", objectInfo.Size, objectInfo.ContentType)

		// 步骤4: 生成下载URL
		t.Log("步骤4: 生成预签名下载URL")
		getURL, err := client.GetPresignedGetURL(ctx, bucket, objectKey, PresignConfig{
			Expires: time.Hour,
		})
		assert.NoError(t, err)
		t.Logf("生成下载URL: %s", getURL.String())

		// 步骤5: 使用预签名URL下载文件
		t.Log("步骤5: 使用预签名URL下载文件")
		downloadResp, err := http.Get(getURL.String())
		assert.NoError(t, err)
		defer downloadResp.Body.Close()

		// 验证下载成功
		assert.Equal(t, http.StatusOK, downloadResp.StatusCode, "下载响应状态码应为200")
		t.Logf("下载成功，状态码: %d", downloadResp.StatusCode)

		// 步骤6: 保存下载的文件
		t.Log("步骤6: 保存下载的文件")
		downloadFile := filepath.Join(downloadDir, "downloaded.txt")
		out, err := os.Create(downloadFile)
		assert.NoError(t, err)
		defer out.Close()

		written, err := io.Copy(out, downloadResp.Body)
		assert.NoError(t, err)
		out.Close() // 确保文件被完全写入
		t.Logf("文件下载到: %s, 大小: %d 字节", downloadFile, written)

		// 步骤7: 验证下载的文件内容
		t.Log("步骤7: 验证下载的文件内容")
		downloadedContent, err := os.ReadFile(downloadFile)
		assert.NoError(t, err)
		assert.Equal(t, uploadContent, downloadedContent, "下载的文件内容与上传的内容不匹配")
		t.Logf("文件内容验证成功，内容长度: %d 字节", len(downloadedContent))

		// 步骤8: 清理测试文件
		t.Log("步骤8: 清理测试文件")
		err = client.RemoveObject(ctx, bucket+"/"+objectKey, minio.RemoveObjectOptions{})
		assert.NoError(t, err)
		t.Logf("测试对象 %s/%s 已从MinIO中删除", bucket, objectKey)
	})

	// 测试预签名URL的元数据和参数
	t.Run("元数据和参数", func(t *testing.T) {
		// 定义文件对象信息
		bucket := "test-bucket"
		objectKey := fmt.Sprintf("metadata-test-%d.txt", time.Now().Unix())

		// 生成带自定义元数据和参数的上传URL
		t.Log("生成带自定义元数据和参数的上传URL")
		putURL, err := client.GetPresignedPutURL(ctx, bucket, objectKey, PresignConfig{
			Expires:     time.Hour,
			ContentType: "application/json",
			Metadata: map[string]string{
				"description": "JSON测试文件",
				"version":     "1.0",
			},
			QueryParams: map[string][]string{
				"response-content-disposition": {"attachment; filename=\"test.json\""},
			},
		})
		assert.NoError(t, err)
		t.Logf("生成上传URL: %s", putURL.String())

		// 上传JSON内容
		jsonContent := []byte(`{"name":"测试","value":123}`)

		// 创建请求，使用内存中的内容
		req, err := http.NewRequest(http.MethodPut, putURL.String(), bytes.NewReader(jsonContent))
		assert.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Content-Length", fmt.Sprintf("%d", len(jsonContent)))

		// 添加元数据头
		req.Header.Set("x-amz-meta-description", "JSON测试文件")
		req.Header.Set("x-amz-meta-version", "1.0")

		t.Logf("设置Content-Length: %d和元数据", len(jsonContent))

		// 发送请求
		resp, err := http.DefaultClient.Do(req)
		assert.NoError(t, err)
		defer resp.Body.Close()

		// 验证上传成功
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		// 验证元数据
		objectInfo, err := client.StatObject(ctx, bucket+"/"+objectKey, minio.StatObjectOptions{})
		assert.NoError(t, err)
		assert.Equal(t, "application/json", objectInfo.ContentType)
		assert.Equal(t, "JSON测试文件", objectInfo.UserMetadata["Description"])
		assert.Equal(t, "1.0", objectInfo.UserMetadata["Version"])
		t.Logf("元数据验证成功: %v", objectInfo.UserMetadata)

		// 清理对象
		err = client.RemoveObject(ctx, bucket+"/"+objectKey, minio.RemoveObjectOptions{})
		assert.NoError(t, err)
	})
}
