package minio

import (
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestClient(t *testing.T) {
	// 使用本地MinIO服务器
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

	// 确保测试存储桶存在
	err = client.ensureBucketExists(ctx, "test-bucket")
	assert.NoError(t, err)

	// 创建临时测试目录
	tempDir := t.TempDir()

	// 测试上传单个文件
	t.Run("PutObject", func(t *testing.T) {
		// 创建测试文件
		testFile := filepath.Join(tempDir, "test-object.txt")
		testContent := []byte("test content")
		err := os.WriteFile(testFile, testContent, 0644)
		assert.NoError(t, err)
		t.Logf("Created test file at: %s", testFile)

		// 打开文件
		file, err := os.Open(testFile)
		assert.NoError(t, err)
		defer file.Close()

		// 获取文件信息
		fileInfo, err := file.Stat()
		assert.NoError(t, err)
		t.Logf("File size: %d bytes", fileInfo.Size())

		// 上传文件
		objectPath := "test-bucket/test-object.txt"
		t.Logf("Uploading file to MinIO: %s", objectPath)
		info, err := client.PutObject(ctx, objectPath, file, fileInfo.Size(), minio.PutObjectOptions{})
		assert.NoError(t, err)
		assert.NotEmpty(t, info.ETag)
		t.Logf("File uploaded successfully. ETag: %s", info.ETag)

		// 验证文件已上传
		t.Logf("Verifying file in MinIO: %s", objectPath)
		obj, err := client.GetObject(ctx, objectPath, minio.GetObjectOptions{})
		assert.NoError(t, err)
		defer obj.Close()

		content, err := io.ReadAll(obj)
		assert.NoError(t, err)
		assert.Equal(t, string(testContent), string(content))
		t.Logf("File verification successful. Content length: %d bytes", len(content))

		// 获取文件信息
		statInfo, err := client.StatObject(ctx, objectPath, minio.StatObjectOptions{})
		assert.NoError(t, err)
		t.Logf("File info in MinIO - Size: %d bytes, Last Modified: %s", statInfo.Size, statInfo.LastModified)
	})

	// 测试批量上传文件
	t.Run("PutObjects", func(t *testing.T) {
		// 创建测试文件
		testFiles := map[string]string{
			"test-bucket/tt/test1.txt": "test1 content",
			"test-bucket/tt/test2.txt": "test2 content",
		}

		objects := make(map[string]io.Reader)
		for path, content := range testFiles {
			// 创建本地文件
			localPath := filepath.Join(tempDir, filepath.Base(path))
			err := os.WriteFile(localPath, []byte(content), 0644)
			assert.NoError(t, err)

			// 打开文件
			file, err := os.Open(localPath)
			assert.NoError(t, err)
			defer file.Close()

			objects[path] = file
		}

		// 批量上传
		err := client.PutObjects(ctx, objects, minio.PutObjectOptions{})
		assert.NoError(t, err)

		// 验证文件已上传
		for path, expectedContent := range testFiles {
			obj, err := client.GetObject(ctx, path, minio.GetObjectOptions{})
			assert.NoError(t, err)
			defer obj.Close()

			content, err := io.ReadAll(obj)
			assert.NoError(t, err)
			assert.Equal(t, expectedContent, string(content))
		}
	})

	// 测试大文件上传
	t.Run("PutLargeObject", func(t *testing.T) {
		// 创建10MB的测试文件
		largeFile := filepath.Join(tempDir, "large-file.txt")
		largeContent := bytes.Repeat([]byte("test"), 2.5*1024*1024) // 10MB
		err := os.WriteFile(largeFile, largeContent, 0644)
		assert.NoError(t, err)

		// 打开文件
		file, err := os.Open(largeFile)
		assert.NoError(t, err)
		defer file.Close()

		// 获取文件信息
		fileInfo, err := file.Stat()
		assert.NoError(t, err)

		// 上传大文件
		info, err := client.PutLargeObject(ctx, "test-bucket/large-file.txt", file, fileInfo.Size(), minio.PutObjectOptions{})
		assert.NoError(t, err)
		assert.NotEmpty(t, info.ETag)

		// 验证大文件已上传
		obj, err := client.GetObject(ctx, "test-bucket/large-file.txt", minio.GetObjectOptions{})
		assert.NoError(t, err)
		defer obj.Close()

		content, err := io.ReadAll(obj)
		assert.NoError(t, err)
		assert.Equal(t, len(largeContent), len(content))
		assert.Equal(t, largeContent, content)
	})

	// 测试列出对象
	t.Run("ListObjects", func(t *testing.T) {
		// 等待一下确保文件上传完成
		time.Sleep(time.Second)

		// 1. 测试列出所有对象（根目录，递归）
		t.Log("测试列出所有对象（根目录，递归）")
		objects := client.ListObjects(ctx, "test-bucket", true)
		count := 0
		objectNames := make(map[string]bool)
		for obj := range objects {
			assert.NotEmpty(t, obj.Object.Key)
			t.Logf("obj.Objecy.Key: %s", obj.Object.Key)
			objectNames[obj.Object.Key] = true
			count++
			t.Logf("Found object: %s, isDir: %v", obj.Object.Key, obj.IsDir)
		}
		assert.Greater(t, count, 0, "应该至少有一个对象")
		assert.True(t, objectNames["test-object.txt"], "应该包含test-object.txt")
		assert.True(t, objectNames["tt/test1.txt"], "应该包含tt/test1.txt")
		assert.True(t, objectNames["tt/test2.txt"], "应该包含tt/test2.txt")
		assert.True(t, objectNames["large-file.txt"], "应该包含large-file.txt")

		// 2. 测试列出特定目录（非递归）
		t.Log("测试列出特定目录（非递归）")
		objects = client.ListObjects(ctx, "test-bucket/tt", false)
		count = 0
		objectNames = make(map[string]bool)
		for obj := range objects {
			assert.NotEmpty(t, obj.Object.Key)
			objectNames[obj.Object.Key] = true
			count++
			t.Logf("Found object in tt/: %s, isDir: %v", obj.Object.Key, obj.IsDir)
		}
		assert.Greater(t, count, 0, "应该至少有一个对象在tt/目录")
		assert.True(t, objectNames["test1.txt"] || objectNames["tt/test1.txt"], "应该包含test1.txt或tt/test1.txt")
		assert.True(t, objectNames["test2.txt"] || objectNames["tt/test2.txt"], "应该包含test2.txt或tt/test2.txt")

		// 3. 测试精确匹配前缀（包含扩展名）
		t.Log("测试精确匹配前缀（包含扩展名）")
		objects = client.ListObjects(ctx, "test-bucket/large-file", true)
		count = 0
		objectNames = make(map[string]bool)
		for obj := range objects {
			assert.NotEmpty(t, obj.Object.Key)
			objectNames[obj.Object.Key] = true
			count++
			t.Logf("Found object with prefix 'large-file': %s, isDir: %v", obj.Object.Key, obj.IsDir)
		}
		assert.Greater(t, count, 0, "应该至少有一个以'large-file'开头的对象")
		assert.True(t, objectNames["large-file.txt"], "应该包含large-file.txt")
	})

	// 测试单个文件下载
	t.Run("GetObject", func(t *testing.T) {
		// 创建下载目录
		downloadDir := filepath.Join(tempDir, "download-single")
		err := os.MkdirAll(downloadDir, 0755)
		assert.NoError(t, err)
		t.Logf("创建下载目录: %s", downloadDir)

		// 1. 获取单个文件
		objectPath := "test-bucket/large-file.txt"
		downloadPath := filepath.Join(downloadDir, "downloaded-large-file.txt")
		t.Logf("获取文件: %s 并下载到: %s", objectPath, downloadPath)

		obj, err := client.GetObject(ctx, objectPath, minio.GetObjectOptions{})
		assert.NoError(t, err)
		defer obj.Close()

		// 2. 获取文件信息
		info, err := obj.Stat()
		assert.NoError(t, err)
		t.Logf("要下载的文件大小: %d bytes", info.Size)

		// 3. 下载文件到本地
		localFile, err := os.Create(downloadPath)
		assert.NoError(t, err)
		defer localFile.Close()

		written, err := io.Copy(localFile, obj)
		assert.NoError(t, err)
		assert.Equal(t, info.Size, written, "下载的数据大小应与文件大小相同")
		localFile.Close() // 确保文件被刷新到磁盘

		// 4. 验证下载的文件
		fileInfo, err := os.Stat(downloadPath)
		assert.NoError(t, err)
		assert.Equal(t, info.Size, fileInfo.Size())
		t.Logf("文件已成功下载到本地，路径: %s，大小: %d bytes", downloadPath, fileInfo.Size())

		// 5. 验证文件内容（部分）
		downloadedContent, err := os.ReadFile(downloadPath)
		assert.NoError(t, err)
		assert.Equal(t, []byte("test"), downloadedContent[:4], "文件内容应该以'test'开头")
	})

	// 测试批量文件下载
	t.Run("GetObjects", func(t *testing.T) {
		// 创建下载目录
		downloadDir := filepath.Join(tempDir, "download-batch")
		err := os.MkdirAll(downloadDir, 0755)
		assert.NoError(t, err)
		t.Logf("创建批量下载目录: %s", downloadDir)

		objectPaths := []string{
			"test-bucket/tt/test1.txt",
			"test-bucket/tt/test2.txt",
		}

		t.Logf("批量获取文件: %v", objectPaths)
		objects, err := client.GetObjects(ctx, objectPaths, minio.GetObjectOptions{})
		assert.NoError(t, err)
		assert.Len(t, objects, 2)

		// 读取并验证每个文件内容，并保存到本地
		expectedContents := map[string]string{
			"test-bucket/tt/test1.txt": "test1 content",
			"test-bucket/tt/test2.txt": "test2 content",
		}

		for path, obj := range objects {
			defer obj.Close()

			// 创建本地文件
			fileName := filepath.Base(path)
			localPath := filepath.Join(downloadDir, fileName)
			t.Logf("下载文件 %s 到: %s", path, localPath)

			localFile, err := os.Create(localPath)
			assert.NoError(t, err)

			// 下载文件内容
			_, err = io.Copy(localFile, obj)
			assert.NoError(t, err)
			localFile.Close()

			// 验证文件大小和内容
			fileContent, err := os.ReadFile(localPath)
			assert.NoError(t, err)
			assert.Equal(t, expectedContents[path], string(fileContent), "文件 %s 内容不匹配", path)
			t.Logf("文件 %s 已成功下载到本地并验证", localPath)
		}
	})

	// 测试大文件下载
	t.Run("GetLargeObject", func(t *testing.T) {
		// 创建下载目录
		downloadDir := filepath.Join(tempDir, "download-large")
		err := os.MkdirAll(downloadDir, 0755)
		assert.NoError(t, err)
		t.Logf("创建大文件下载目录: %s", downloadDir)

		objectPath := "test-bucket/large-file.txt"
		downloadPath := filepath.Join(downloadDir, "large-file-download.txt")
		t.Logf("分片下载大文件: %s 到 %s", objectPath, downloadPath)

		// 创建本地文件
		localFile, err := os.Create(downloadPath)
		assert.NoError(t, err)
		defer localFile.Close()

		// 使用GetLargeObject下载
		err = client.GetLargeObject(ctx, objectPath, localFile, minio.GetObjectOptions{})
		assert.NoError(t, err)
		localFile.Close() // 确保文件被刷新到磁盘

		// 验证下载的文件
		fileInfo, err := os.Stat(downloadPath)
		assert.NoError(t, err)
		assert.Equal(t, int64(10*1024*1024), fileInfo.Size(), "下载的文件大小不匹配")
		t.Logf("成功下载大文件到本地路径: %s，大小: %d bytes", downloadPath, fileInfo.Size())

		// 验证文件内容（部分检查）
		content, err := os.ReadFile(downloadPath)
		assert.NoError(t, err)
		assert.Equal(t, []byte("test"), content[:4], "文件内容应该以'test'开头")
		t.Logf("大文件内容验证成功")
	})

	// 测试删除对象
	t.Run("RemoveObject", func(t *testing.T) {
		// 先确保文件存在
		_, err := client.StatObject(ctx, "test-bucket/test-object.txt", minio.StatObjectOptions{})
		assert.NoError(t, err, "文件应该存在")

		// 删除文件
		err = client.RemoveObject(ctx, "test-bucket/test-object.txt", minio.RemoveObjectOptions{})
		assert.NoError(t, err)

		// 验证文件已删除
		_, err = client.StatObject(ctx, "test-bucket/test-object.txt", minio.StatObjectOptions{})
		assert.Error(t, err, "文件应该已被删除")
	})

	// 测试批量删除对象
	t.Run("RemoveObjects", func(t *testing.T) {
		objectPaths := []string{
			"test-bucket/tt/test1.txt",
			"test-bucket/tt/test2.txt",
		}

		// 先确保文件存在
		for _, path := range objectPaths {
			_, err := client.StatObject(ctx, path, minio.StatObjectOptions{})
			assert.NoError(t, err, "文件应该存在: "+path)
		}

		// 删除文件
		err := client.RemoveObjects(ctx, objectPaths, minio.RemoveObjectOptions{})
		assert.NoError(t, err)

		// 验证文件已删除
		for _, path := range objectPaths {
			_, err := client.StatObject(ctx, path, minio.StatObjectOptions{})
			assert.Error(t, err, "文件应该已被删除: "+path)
		}
	})

	// 测试获取对象信息
	t.Run("StatObject", func(t *testing.T) {
		info, err := client.StatObject(ctx, "test-bucket/large-file.txt", minio.StatObjectOptions{})
		assert.NoError(t, err)
		assert.Equal(t, int64(10*1024*1024), info.Size)
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

		// 验证文件已复制
		obj, err := client.GetObject(ctx, "test-bucket/large-file-copy.txt", minio.GetObjectOptions{})
		assert.NoError(t, err)
		defer obj.Close()

		content, err := io.ReadAll(obj)
		assert.NoError(t, err)
		assert.Equal(t, 10*1024*1024, len(content))
	})

	// 测试健康检查
	t.Run("HealthCheck", func(t *testing.T) {
		err := client.HealthCheck(ctx)
		assert.NoError(t, err)
	})
}
