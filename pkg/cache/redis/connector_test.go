package redis

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testRedisAddr     = "localhost:6379"
	testRedisPassword = ""
	testRedisDB       = 0
	testRedisPoolSize = 10
)

func TestRedisConnector(t *testing.T) {
	ctx := context.Background()

	t.Run("连接创建", func(t *testing.T) {
		connector, err := NewConnector(ctx, testRedisAddr, testRedisPassword, testRedisDB, testRedisPoolSize)
		require.NoError(t, err)
		require.NotNil(t, connector)
		defer connector.Close()

		// 验证连接有效
		assert.NoError(t, connector.HealthCheck())

		// 验证客户端可用
		client := connector.GetClient()
		assert.NotNil(t, client)

		// 测试设置和获取值
		key := "test:connector:key"
		value := "hello world"
		err = client.Set(ctx, key, value, time.Minute).Err()
		assert.NoError(t, err)

		// 获取值并验证
		result, err := client.Get(ctx, key).Result()
		assert.NoError(t, err)
		assert.Equal(t, value, result)

		// 清理测试键
		err = client.Del(ctx, key).Err()
		assert.NoError(t, err)
	})

	t.Run("连接错误", func(t *testing.T) {
		// 测试无效连接
		_, err := NewConnector(ctx, "invalid-host:6379", "", 0, 10)
		assert.Error(t, err)
		assert.Equal(t, ErrConnFailed, err)
	})
}
