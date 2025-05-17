package redis

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDistributedLock(t *testing.T) {
	ctx := context.Background()

	// 创建连接
	connector, err := NewConnector(ctx, testRedisAddr, testRedisPassword, testRedisDB, testRedisPoolSize)
	require.NoError(t, err)
	require.NotNil(t, connector)
	defer connector.Close()

	client := connector.GetClient()

	t.Run("锁获取和释放", func(t *testing.T) {
		lockKey := "test:lock:basic"
		ttl := time.Second * 30

		// 清理可能存在的锁
		client.Del(ctx, lockKey)

		// 创建锁实例
		lock := NewDistributedLock(ctx, client, lockKey, ttl)

		// 测试获取锁
		err := lock.Acquire()
		assert.NoError(t, err)

		// 验证锁已设置
		exists, err := client.Exists(ctx, lockKey).Result()
		assert.NoError(t, err)
		assert.Equal(t, int64(1), exists)

		// 测试释放锁
		err = lock.Release()
		assert.NoError(t, err)

		// 验证锁已释放
		exists, err = client.Exists(ctx, lockKey).Result()
		assert.NoError(t, err)
		assert.Equal(t, int64(0), exists)
	})

	t.Run("锁互斥性", func(t *testing.T) {
		lockKey := "test:lock:exclusivity"
		ttl := time.Second * 5

		// 清理可能存在的锁
		client.Del(ctx, lockKey)

		// 创建第一个锁并获取
		lock1 := NewDistributedLock(ctx, client, lockKey, ttl)
		err := lock1.Acquire()
		assert.NoError(t, err)

		// 创建第二个锁尝试获取同一个锁
		lock2 := NewDistributedLock(ctx, client, lockKey, ttl)
		err = lock2.Acquire()
		assert.Error(t, err)
		assert.Equal(t, ErrLockNotAcquired, err)

		// 释放第一个锁
		err = lock1.Release()
		assert.NoError(t, err)

		// 现在第二个锁应该可以获取
		err = lock2.Acquire()
		assert.NoError(t, err)

		// 清理
		lock2.Release()
	})

	t.Run("锁续期", func(t *testing.T) {
		lockKey := "test:lock:refresh"
		ttl := time.Second * 2 // 短暂的过期时间

		// 清理可能存在的锁
		client.Del(ctx, lockKey)

		// 创建锁并获取
		lock := NewDistributedLock(ctx, client, lockKey, ttl)
		err := lock.Acquire()
		assert.NoError(t, err)

		// 延迟1秒，然后刷新锁
		time.Sleep(time.Second)
		err = lock.Refresh()
		assert.NoError(t, err)

		// 获取剩余TTL，应该接近初始TTL
		ttlVal, err := client.TTL(ctx, lockKey).Result()
		assert.NoError(t, err)
		assert.True(t, ttlVal > time.Second, "锁续期应该延长TTL")

		// 清理
		lock.Release()
	})

	t.Run("并发锁竞争", func(t *testing.T) {
		lockKey := "test:lock:concurrent"
		ttl := time.Second * 5

		// 清理可能存在的锁
		client.Del(ctx, lockKey)

		// 记录成功获取锁的协程数量
		var acquired int
		var mu sync.Mutex

		// 并发获取锁
		var wg sync.WaitGroup
		for i := 0; i < 5; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()

				lock := NewDistributedLock(ctx, client, lockKey, ttl)
				err := lock.Acquire()

				if err == nil {
					mu.Lock()
					acquired++
					mu.Unlock()

					// 模拟处理时间
					time.Sleep(100 * time.Millisecond)

					// 释放锁
					lock.Release()
				}
			}()
		}

		wg.Wait()

		// 验证只有一个协程能成功获取锁
		assert.Equal(t, 1, acquired, "在并发环境下，只有一个协程应该能获取锁")
	})
}
