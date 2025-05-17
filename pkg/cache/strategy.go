package cache

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/blackarbiter/go-sac/pkg/cache/redis"
	goredis "github.com/go-redis/redis/v8"
)

// 定义标准错误类型
var (
	ErrCacheMiss       = errors.New("cache: key not found")
	ErrInvalidType     = errors.New("cache: invalid value type")
	ErrSerialization   = errors.New("cache: serialization failed")
	ErrDeserialization = errors.New("cache: deserialization failed")
)

// CacheStrategy 统一缓存策略接口
type CacheStrategy interface {
	Get(ctx context.Context, key string, target interface{}) error
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	Delete(ctx context.Context, keys ...string) error
	Exists(ctx context.Context, key string) (bool, error)
	TTL(ctx context.Context, key string) (time.Duration, error)
}

// RedisCacheStrategy Redis缓存实现
type RedisCacheStrategy struct {
	conn        *redis.Connector
	fallback    CacheStrategy // 多级缓存回退
	compression bool          // 是否启用压缩
}

// NewRedisCacheStrategy 创建Redis缓存策略实例
func NewRedisCacheStrategy(conn *redis.Connector) *RedisCacheStrategy {
	return &RedisCacheStrategy{
		conn: conn,
	}
}

// Get 获取缓存值（带自动解压和回退逻辑）
func (r *RedisCacheStrategy) Get(ctx context.Context, key string, target interface{}) error {
	// 从Redis获取原始数据
	data, err := r.conn.GetClient().Get(ctx, key).Bytes()
	if err != nil {
		if errors.Is(err, goredis.Nil) && r.fallback != nil {
			// 回退到次级缓存
			return r.fallback.Get(ctx, key, target)
		}
		return fmt.Errorf("%w: %v", ErrCacheMiss, err)
	}

	// 处理压缩数据
	if r.compression {
		data, err = decompress(data)
		if err != nil {
			return fmt.Errorf("%w: %v", ErrDeserialization, err)
		}
	}

	// 反序列化数据
	if err := json.Unmarshal(data, target); err != nil {
		return fmt.Errorf("%w: %v", ErrDeserialization, err)
	}
	return nil
}

// Set 设置缓存值（带压缩和序列化）
func (r *RedisCacheStrategy) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	// 序列化数据
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrSerialization, err)
	}

	// 压缩处理
	if r.compression {
		data, err = compress(data)
		if err != nil {
			return fmt.Errorf("compress error: %w", err)
		}
	}

	// 存储到Redis
	if err := r.conn.GetClient().Set(ctx, key, data, ttl).Err(); err != nil {
		return fmt.Errorf("redis set error: %w", err)
	}

	// 同步到回退缓存
	if r.fallback != nil {
		_ = r.fallback.Set(ctx, key, value, ttl) // 忽略次级缓存错误
	}
	return nil
}

// Delete 批量删除键
func (r *RedisCacheStrategy) Delete(ctx context.Context, keys ...string) error {
	if len(keys) == 0 {
		return nil
	}

	if err := r.conn.GetClient().Del(ctx, keys...).Err(); err != nil {
		return fmt.Errorf("redis delete error: %w", err)
	}

	// 清理回退缓存
	if r.fallback != nil {
		_ = r.fallback.Delete(ctx, keys...) // 忽略次级缓存错误
	}
	return nil
}

// Exists 检查键是否存在
func (r *RedisCacheStrategy) Exists(ctx context.Context, key string) (bool, error) {
	count, err := r.conn.GetClient().Exists(ctx, key).Result()
	return count > 0, err
}

// TTL 获取剩余时间
func (r *RedisCacheStrategy) TTL(ctx context.Context, key string) (time.Duration, error) {
	return r.conn.GetClient().TTL(ctx, key).Result()
}

// WithFallback 设置多级缓存回退
func (r *RedisCacheStrategy) WithFallback(fallback CacheStrategy) *RedisCacheStrategy {
	r.fallback = fallback
	return r
}

// WithCompression 启用GZIP压缩
func (r *RedisCacheStrategy) WithCompression(enable bool) *RedisCacheStrategy {
	r.compression = enable
	return r
}

// 压缩工具函数
func compress(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	if _, err := gw.Write(data); err != nil {
		return nil, err
	}
	if err := gw.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// 解压工具函数
func decompress(data []byte) ([]byte, error) {
	gr, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer gr.Close()

	return io.ReadAll(gr)
}
