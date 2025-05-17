// pkg/storage/mysql/connector.go
package mysql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"go.uber.org/zap"
)

var (
	// ErrMaxRetriesExceeded 表示连接尝试超出最大重试次数
	ErrMaxRetriesExceeded = errors.New("超出最大重试次数")
)

type ConnectorConfig struct {
	DSN             string        // 数据库连接字符串
	MaxOpenConns    int           // 最大打开连接数
	MaxIdleConns    int           // 最大空闲连接数
	ConnMaxLifetime time.Duration // 连接最大生命周期
	ConnTimeout     time.Duration // 连接超时
	RetryAttempts   int           // 重试次数
	RetryDelay      time.Duration // 重试延迟
}

type DBConnector struct {
	db     *sql.DB
	logger *zap.Logger
	config ConnectorConfig
}

func NewConnector(cfg ConnectorConfig, logger *zap.Logger) (*DBConnector, error) {
	db, err := sql.Open("mysql", cfg.DSN)
	if err != nil {
		logger.Error("数据库连接初始化失败",
			zap.String("dsn", cfg.DSN),
			zap.Error(err))
		return nil, fmt.Errorf("数据库连接初始化失败: %w", err)
	}

	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	// 使用带超时的上下文测试连接
	ctx, cancel := context.WithTimeout(context.Background(), cfg.ConnTimeout)
	defer cancel()

	// 添加重试逻辑
	err = retry(ctx, cfg.RetryAttempts, cfg.RetryDelay, func() error {
		return db.PingContext(ctx)
	})

	if err != nil {
		logger.Error("数据库连接测试失败",
			zap.String("dsn", cfg.DSN),
			zap.Error(err))
		return nil, fmt.Errorf("数据库连接测试失败: %w", err)
	}

	logger.Info("数据库连接池初始化成功",
		zap.String("dsn", cfg.DSN),
		zap.Int("max_open_conns", cfg.MaxOpenConns),
		zap.Int("max_idle_conns", cfg.MaxIdleConns))

	return &DBConnector{
		db:     db,
		logger: logger.Named("mysql.connector"),
		config: cfg,
	}, nil
}

func (c *DBConnector) GetDB() *sql.DB {
	return c.db
}

func (c *DBConnector) Close() error {
	c.logger.Info("正在关闭数据库连接池...")
	if err := c.db.Close(); err != nil {
		c.logger.Error("关闭数据库连接池失败",
			zap.Error(err))
		return fmt.Errorf("关闭数据库连接池失败: %w", err)
	}
	return nil
}

// 带重试的操作
func retry(ctx context.Context, attempts int, delay time.Duration, fn func() error) error {
	if attempts <= 0 {
		attempts = 1
	}

	var err error
	for i := 0; i < attempts; i++ {
		err = fn()
		if err == nil {
			return nil
		}

		// 如果上下文已取消，立即返回
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			// 继续执行
		}

		// 最后一次尝试不需要等待
		if i < attempts-1 {
			time.Sleep(delay)
		}
	}

	if err != nil {
		return fmt.Errorf("%w: %v", ErrMaxRetriesExceeded, err)
	}

	return nil
}

// WithTransaction 执行事务操作
func (c *DBConnector) WithTransaction(ctx context.Context, fn func(*sql.Tx) error) error {
	tx, err := c.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("开始事务失败: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			// 发生panic，回滚事务
			tx.Rollback()
			panic(p) // 重新抛出panic
		}
	}()

	if err := fn(tx); err != nil {
		// 发生错误，回滚事务
		if rbErr := tx.Rollback(); rbErr != nil {
			c.logger.Error("事务回滚失败",
				zap.Error(rbErr))
			return fmt.Errorf("执行事务失败: %v, 回滚失败: %v", err, rbErr)
		}
		return err
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		c.logger.Error("事务提交失败",
			zap.Error(err))
		return fmt.Errorf("事务提交失败: %w", err)
	}

	return nil
}

// ExecContext 安全执行SQL，带重试机制
func (c *DBConnector) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	var result sql.Result
	err := retry(ctx, c.config.RetryAttempts, c.config.RetryDelay, func() error {
		var err error
		result, err = c.db.ExecContext(ctx, query, args...)
		return err
	})
	return result, err
}

// QueryContext 安全查询SQL，带重试机制
func (c *DBConnector) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	var rows *sql.Rows
	err := retry(ctx, c.config.RetryAttempts, c.config.RetryDelay, func() error {
		var err error
		rows, err = c.db.QueryContext(ctx, query, args...)
		return err
	})
	return rows, err
}

// QueryRowContext 安全查询单行SQL，带重试机制
func (c *DBConnector) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	return c.db.QueryRowContext(ctx, query, args...)
}
