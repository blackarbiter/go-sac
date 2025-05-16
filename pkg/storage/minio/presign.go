package minio

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

// PresignConfig 预签名配置
type PresignConfig struct {
	Expires     time.Duration     // URL有效期
	ContentType string            // 内容类型
	Metadata    map[string]string // 自定义元数据
	QueryParams url.Values        // 自定义查询参数
}

// GetPresignedPutURL 生成临时上传URL
func (c *Client) GetPresignedPutURL(ctx context.Context, bucket, object string, cfg PresignConfig) (*url.URL, error) {
	// 设置请求参数
	reqParams := cfg.QueryParams
	if reqParams == nil {
		reqParams = make(url.Values)
	}

	// 创建请求头
	headers := make(http.Header)
	if cfg.ContentType != "" {
		headers.Set("Content-Type", cfg.ContentType)
	}

	// 添加元数据
	for k, v := range cfg.Metadata {
		headers.Set("x-amz-meta-"+k, v)
	}

	// 设置过期时间，如果未设置，默认为24小时
	expires := cfg.Expires
	if expires == 0 {
		expires = 24 * time.Hour
	}

	// 直接使用Minio客户端的预签名功能
	presignedURL, err := c.client.PresignHeader(ctx, "PUT", bucket, object, expires, reqParams, headers)
	if err != nil {
		return nil, fmt.Errorf("生成上传预签名URL失败: %w", err)
	}

	return presignedURL, nil
}

// GetPresignedGetURL 生成临时下载URL
func (c *Client) GetPresignedGetURL(ctx context.Context, bucket, object string, cfg PresignConfig) (*url.URL, error) {
	// 设置请求参数
	reqParams := cfg.QueryParams
	if reqParams == nil {
		reqParams = make(url.Values)
	}

	// 设置过期时间，如果未设置，默认为24小时
	expires := cfg.Expires
	if expires == 0 {
		expires = 24 * time.Hour
	}

	// 创建请求头（对于GET请求，通常不需要额外请求头）
	headers := make(http.Header)
	for k, v := range cfg.Metadata {
		headers.Set(k, v)
	}

	// 直接使用Minio客户端的预签名功能
	presignedURL, err := c.client.PresignHeader(ctx, "GET", bucket, object, expires, reqParams, headers)
	if err != nil {
		return nil, fmt.Errorf("生成下载预签名URL失败: %w", err)
	}

	return presignedURL, nil
}
