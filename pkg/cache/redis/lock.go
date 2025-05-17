package redis

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"time"

	"github.com/go-redis/redis/v8"
)

var (
	ErrLockNotAcquired = errors.New("failed to acquire lock")
	ErrLockNotHeld     = errors.New("lock not held by this client")
)

// DistributedLock 分布式锁结构
type DistributedLock struct {
	client *redis.Client
	key    string
	value  string
	ttl    time.Duration
	ctx    context.Context
}

// NewDistributedLock 创建分布式锁实例
func NewDistributedLock(ctx context.Context, client *redis.Client, key string, ttl time.Duration) *DistributedLock {
	return &DistributedLock{
		client: client,
		key:    key,
		ttl:    ttl,
		ctx:    ctx,
	}
}

// generateLockValue 生成唯一锁标识
func (dl *DistributedLock) generateLockValue() (string, error) {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(b), nil
}

// Acquire 获取分布式锁
func (dl *DistributedLock) Acquire() error {
	val, err := dl.generateLockValue()
	if err != nil {
		return err
	}

	result, err := dl.client.SetNX(dl.ctx, dl.key, val, dl.ttl).Result()
	if err != nil {
		return err
	}
	if !result {
		return ErrLockNotAcquired
	}
	dl.value = val
	return nil
}

// Release 释放分布式锁
func (dl *DistributedLock) Release() error {
	script := `
	if redis.call("get", KEYS[1]) == ARGV[1] then
		return redis.call("del", KEYS[1])
	else
		return 0
	end`

	_, err := dl.client.Eval(dl.ctx, script, []string{dl.key}, dl.value).Result()
	if err == redis.Nil {
		return ErrLockNotHeld
	}
	return err
}

// Refresh 续期锁有效期
func (dl *DistributedLock) Refresh() error {
	ok, err := dl.client.Expire(dl.ctx, dl.key, dl.ttl).Result()
	if err != nil {
		return err
	}
	if !ok {
		return ErrLockNotHeld
	}
	return nil
}
