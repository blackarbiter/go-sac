package cache

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/blackarbiter/go-sac/pkg/cache/redis"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testRedisAddr     = "localhost:6379"
	testRedisPassword = ""
	testRedisDB       = 1 // 使用不同的DB避免与其他测试冲突
	testRedisPoolSize = 10
)

// 测试用结构
type TestUser struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	CreateAt int64  `json:"create_at"`
}

func TestRedisCacheStrategy(t *testing.T) {
	ctx := context.Background()

	// 创建Redis连接
	connector, err := redis.NewConnector(ctx, testRedisAddr, testRedisPassword, testRedisDB, testRedisPoolSize)
	require.NoError(t, err)
	defer connector.Close()

	// 创建缓存策略
	cacheStrategy := NewRedisCacheStrategy(connector)

	// 测试数据
	testKey := "test:user:1"
	testUser := TestUser{
		ID:       1,
		Name:     "测试用户",
		Email:    "test@example.com",
		CreateAt: time.Now().Unix(),
	}

	t.Run("基本缓存功能", func(t *testing.T) {
		// 确保初始状态无缓存
		cacheStrategy.Delete(ctx, testKey)

		// 检查键不存在
		exists, err := cacheStrategy.Exists(ctx, testKey)
		assert.NoError(t, err)
		assert.False(t, exists)

		// 设置缓存
		err = cacheStrategy.Set(ctx, testKey, testUser, time.Minute*5)
		assert.NoError(t, err)

		// 验证键存在
		exists, err = cacheStrategy.Exists(ctx, testKey)
		assert.NoError(t, err)
		assert.True(t, exists)

		// 获取TTL
		ttl, err := cacheStrategy.TTL(ctx, testKey)
		assert.NoError(t, err)
		assert.True(t, ttl > 0)

		// 获取缓存
		var retrievedUser TestUser
		err = cacheStrategy.Get(ctx, testKey, &retrievedUser)
		assert.NoError(t, err)
		assert.Equal(t, testUser.ID, retrievedUser.ID)
		assert.Equal(t, testUser.Name, retrievedUser.Name)
		assert.Equal(t, testUser.Email, retrievedUser.Email)

		// 删除缓存
		err = cacheStrategy.Delete(ctx, testKey)
		assert.NoError(t, err)

		// 验证键不存在
		exists, err = cacheStrategy.Exists(ctx, testKey)
		assert.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("压缩功能测试", func(t *testing.T) {
		// 启用压缩功能
		compressedCache := cacheStrategy.WithCompression(true)

		// 清理可能存在的数据
		compressedCache.Delete(ctx, testKey)

		// 创建大量数据
		largeUser := TestUser{
			ID:       2,
			Name:     "大数据用户",
			Email:    "large@example.com",
			CreateAt: time.Now().Unix(),
		}

		// 设置压缩缓存
		err = compressedCache.Set(ctx, testKey, largeUser, time.Minute*5)
		assert.NoError(t, err)

		// 读取压缩数据
		var retrievedLargeUser TestUser
		err = compressedCache.Get(ctx, testKey, &retrievedLargeUser)
		assert.NoError(t, err)
		assert.Equal(t, largeUser.ID, retrievedLargeUser.ID)
		assert.Equal(t, largeUser.Name, retrievedLargeUser.Name)

		// 清理
		compressedCache.Delete(ctx, testKey)
	})

	t.Run("多级缓存回退", func(t *testing.T) {
		// 创建内存缓存模拟
		mockCache := &mockCacheStrategy{
			data: make(map[string][]byte),
		}

		// 创建带回退的缓存
		fallbackCache := cacheStrategy.WithFallback(mockCache)
		fallbackKey := "test:fallback:key"

		// 确保Redis中不存在此键
		fallbackCache.Delete(ctx, fallbackKey)

		// 设置数据到回退缓存
		fallbackUser := TestUser{
			ID:    3,
			Name:  "回退用户",
			Email: "fallback@example.com",
		}
		err = mockCache.Set(ctx, fallbackKey, fallbackUser, time.Minute)
		assert.NoError(t, err)

		// 从主缓存获取，预期会回退到mockCache
		var retrievedFallbackUser TestUser
		err = fallbackCache.Get(ctx, fallbackKey, &retrievedFallbackUser)
		assert.NoError(t, err)
		assert.Equal(t, fallbackUser.ID, retrievedFallbackUser.ID)
		assert.Equal(t, fallbackUser.Name, retrievedFallbackUser.Name)

		// 设置主缓存，验证是否同步到回退缓存
		primaryUser := TestUser{
			ID:    4,
			Name:  "主缓存用户",
			Email: "primary@example.com",
		}
		err = fallbackCache.Set(ctx, fallbackKey, primaryUser, time.Minute)
		assert.NoError(t, err)

		// 验证回退缓存是否也被更新
		var mockCacheUser TestUser
		err = mockCache.Get(ctx, fallbackKey, &mockCacheUser)
		assert.NoError(t, err)
		assert.Equal(t, primaryUser.ID, mockCacheUser.ID)

		// 清理
		fallbackCache.Delete(ctx, fallbackKey)
	})
}

// 内存缓存实现，模拟回退缓存
type mockCacheStrategy struct {
	data map[string][]byte
}

func (m *mockCacheStrategy) Get(ctx context.Context, key string, target interface{}) error {
	data, ok := m.data[key]
	if !ok {
		return ErrCacheMiss
	}
	return json.Unmarshal(data, target)
}

func (m *mockCacheStrategy) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	m.data[key] = data
	return nil
}

func (m *mockCacheStrategy) Delete(ctx context.Context, keys ...string) error {
	for _, key := range keys {
		delete(m.data, key)
	}
	return nil
}

func (m *mockCacheStrategy) Exists(ctx context.Context, key string) (bool, error) {
	_, ok := m.data[key]
	return ok, nil
}

func (m *mockCacheStrategy) TTL(ctx context.Context, key string) (time.Duration, error) {
	if _, ok := m.data[key]; !ok {
		return -2 * time.Second, nil
	}
	return time.Minute, nil // 模拟固定的TTL
}
