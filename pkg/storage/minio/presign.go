package minio

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/url"
	"time"

	_ "github.com/minio/minio-go/v7/pkg/signer"
)

// PresignConfig 预签名配置
type PresignConfig struct {
	Expires time.Duration
	Secure  bool
	Headers http.Header
}

// GetPresignedPutURL 生成安全的上传URL
func (c *Client) GetPresignedPutURL(ctx context.Context, bucket, object string, cfg PresignConfig) (*url.URL, error) {
	reqParams := make(url.Values)

	presignedURL, err := c.client.PresignHeader(ctx, "PUT", bucket, object, cfg.Expires, reqParams, cfg.Headers)
	if err != nil {
		return nil, fmt.Errorf("生成预签名URL失败: %w", err)
	}

	// 添加安全校验参数
	return c.signWithHmac(presignedURL, cfg), nil
}

// GetPresignedGetURL 生成安全的下载URL
func (c *Client) GetPresignedGetURL(ctx context.Context, bucket, object string, cfg PresignConfig) (*url.URL, error) {
	reqParams := make(url.Values)

	// 使用正确的Presign方法签名
	presignedURL, err := c.client.PresignedGetObject(
		ctx,
		bucket,
		object,
		cfg.Expires,
		reqParams,
	)
	if err != nil {
		return nil, fmt.Errorf("生成预签名URL失败: %w", err)
	}

	return c.signWithHmac(presignedURL, cfg), nil
}

// signWithHmac 增加HMAC二次签名
func (c *Client) signWithHmac(u *url.URL, cfg PresignConfig) *url.URL {
	// 计算HMAC签名
	h := hmac.New(sha256.New, []byte(c.config.SecretKey))
	h.Write([]byte(u.String()))
	signature := hex.EncodeToString(h.Sum(nil))

	// 添加签名参数
	q := u.Query()
	q.Set("x-security-sig", signature)
	u.RawQuery = q.Encode()

	return u
}

// VerifyPresignedRequest 验证请求签名
func (c *Client) VerifyPresignedRequest(r *http.Request) bool {
	if r == nil || r.URL == nil {
		return false
	}

	// 获取URL查询参数的副本，避免修改原始请求
	params := r.URL.Query()

	// 提取签名参数
	signature := params.Get("x-security-sig")
	if signature == "" {
		return false
	}

	// 创建用于计算签名的URL副本
	verifyURL := *r.URL
	verifyParams := verifyURL.Query()
	verifyParams.Del("x-security-sig")
	verifyURL.RawQuery = verifyParams.Encode()

	// 重新计算签名
	h := hmac.New(sha256.New, []byte(c.config.SecretKey))
	h.Write([]byte(verifyURL.String()))
	expected := hex.EncodeToString(h.Sum(nil))

	return hmac.Equal([]byte(signature), []byte(expected))
}
