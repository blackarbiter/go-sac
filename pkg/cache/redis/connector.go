package redis

import (
	"context"
	"errors"
	"time"

	"github.com/go-redis/redis/v8"
)

var (
	ErrConnFailed = errors.New("failed to connect to redis")
)

// Connector Redis连接管理器
type Connector struct {
	client     *redis.Client
	ctx        context.Context
	maxRetries int
}

// NewConnector 创建Redis连接实例
func NewConnector(ctx context.Context, addr, password string, db, poolSize int) (*Connector, error) {
	client := redis.NewClient(&redis.Options{
		Addr:         addr,
		Password:     password,
		DB:           db,
		PoolSize:     poolSize,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	})

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, ErrConnFailed
	}

	return &Connector{
		client:     client,
		ctx:        ctx,
		maxRetries: 3,
	}, nil
}

// GetClient 获取原生客户端实例
func (c *Connector) GetClient() *redis.Client {
	return c.client
}

// HealthCheck 连接健康检查
func (c *Connector) HealthCheck() error {
	return c.client.Ping(c.ctx).Err()
}

// Close 安全关闭连接
func (c *Connector) Close() error {
	return c.client.Close()
}
