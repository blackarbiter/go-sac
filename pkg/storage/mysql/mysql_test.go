package mysql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

const (
	// 测试数据库配置，需要Docker中MySQL正在运行
	testDBHost     = "localhost"
	testDBPort     = 3306
	testDBUser     = "root"
	testDBPassword = "1234qwer"
	testDBName     = "sac_test"
)

// 设置测试环境
func setupTestDatabase(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	// 创建root连接用于初始化测试数据库
	rootDSN := fmt.Sprintf("%s:%s@tcp(%s:%d)/",
		testDBUser, testDBPassword, testDBHost, testDBPort)

	db, err := sql.Open("mysql", rootDSN)
	require.NoError(t, err)
	defer db.Close()

	// 确保测试数据库存在并干净
	_, err = db.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS %s", testDBName))
	require.NoError(t, err)

	_, err = db.Exec(fmt.Sprintf("CREATE DATABASE %s", testDBName))
	require.NoError(t, err)

	t.Logf("测试数据库 %s 已创建", testDBName)
}

// 获取测试DSN
func getTestDSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true&multiStatements=true",
		testDBUser, testDBPassword, testDBHost, testDBPort, testDBName)
}

// 创建临时迁移文件
func setupMigrationFiles(t *testing.T) string {
	// 创建临时目录存放迁移文件
	migrationsDir, err := os.MkdirTemp("", "mysql-migrations")
	require.NoError(t, err)

	// 创建迁移文件
	migrationUp := []byte(`
		CREATE TABLE users (
			id INT AUTO_INCREMENT PRIMARY KEY,
			username VARCHAR(50) NOT NULL,
			email VARCHAR(100) NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
		
		CREATE TABLE posts (
			id INT AUTO_INCREMENT PRIMARY KEY,
			title VARCHAR(100) NOT NULL,
			content TEXT,
			user_id INT,
			FOREIGN KEY (user_id) REFERENCES users(id)
		);
	`)

	migrationDown := []byte(`
		DROP TABLE IF EXISTS posts;
		DROP TABLE IF EXISTS users;
	`)

	err = os.WriteFile(filepath.Join(migrationsDir, "000001_create_tables.up.sql"), migrationUp, 0644)
	require.NoError(t, err)

	err = os.WriteFile(filepath.Join(migrationsDir, "000001_create_tables.down.sql"), migrationDown, 0644)
	require.NoError(t, err)

	return migrationsDir
}

// 测试连接器
func TestConnector(t *testing.T) {
	// 跳过测试如果没有Docker MySQL实例
	if testing.Short() {
		t.Skip("跳过集成测试，需要MySQL实例")
	}

	setupTestDatabase(t)
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	// 测试配置
	cfg := ConnectorConfig{
		DSN:             getTestDSN(),
		MaxOpenConns:    10,
		MaxIdleConns:    5,
		ConnMaxLifetime: time.Minute * 10,
		ConnTimeout:     time.Second * 5,
		RetryAttempts:   3,
		RetryDelay:      time.Millisecond * 100,
	}

	// 测试创建连接器
	t.Run("创建连接器", func(t *testing.T) {
		connector, err := NewConnector(cfg, logger)
		require.NoError(t, err)
		defer connector.Close()

		assert.NotNil(t, connector.GetDB())

		// 测试连接是否正常工作
		err = connector.GetDB().Ping()
		assert.NoError(t, err)
	})

	// 测试事务管理和重试功能
	t.Run("事务管理", func(t *testing.T) {
		connector, err := NewConnector(cfg, logger)
		require.NoError(t, err)
		defer connector.Close()

		// 创建测试表
		_, err = connector.GetDB().Exec(`
			CREATE TABLE test_tx (
				id INT AUTO_INCREMENT PRIMARY KEY,
				name VARCHAR(50) NOT NULL
			)
		`)
		require.NoError(t, err)

		// 测试事务提交
		ctx := context.Background()
		err = connector.WithTransaction(ctx, func(tx *sql.Tx) error {
			_, err := tx.Exec("INSERT INTO test_tx (name) VALUES (?)", "tx-test-1")
			if err != nil {
				return err
			}

			_, err = tx.Exec("INSERT INTO test_tx (name) VALUES (?)", "tx-test-2")
			return err
		})
		require.NoError(t, err)

		// 验证数据已提交
		var count int
		err = connector.GetDB().QueryRow("SELECT COUNT(*) FROM test_tx").Scan(&count)
		assert.NoError(t, err)
		assert.Equal(t, 2, count)

		// 测试事务回滚
		err = connector.WithTransaction(ctx, func(tx *sql.Tx) error {
			_, err := tx.Exec("INSERT INTO test_tx (name) VALUES (?)", "tx-test-3")
			if err != nil {
				return err
			}

			return errors.New("模拟事务失败")
		})
		assert.Error(t, err)

		// 验证数据已回滚
		err = connector.GetDB().QueryRow("SELECT COUNT(*) FROM test_tx").Scan(&count)
		assert.NoError(t, err)
		assert.Equal(t, 2, count, "事务应该已回滚，计数应该不变")

		// 测试带重试的操作
		var result sql.Result
		result, err = connector.ExecContext(ctx, "INSERT INTO test_tx (name) VALUES (?)", "retry-test")
		assert.NoError(t, err)

		affected, err := result.RowsAffected()
		assert.NoError(t, err)
		assert.Equal(t, int64(1), affected)
	})
}

// 测试监控器
func TestMonitor(t *testing.T) {
	// 跳过测试如果没有Docker MySQL实例
	if testing.Short() {
		t.Skip("跳过集成测试，需要MySQL实例")
	}

	setupTestDatabase(t)
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	// 创建连接器
	connector, err := NewConnector(ConnectorConfig{
		DSN:             getTestDSN(),
		MaxOpenConns:    10,
		MaxIdleConns:    5,
		ConnMaxLifetime: time.Minute * 10,
		ConnTimeout:     time.Second * 5,
	}, logger)
	require.NoError(t, err)
	defer connector.Close()

	// 创建表
	_, err = connector.GetDB().Exec(`
		CREATE TABLE test_monitor (
			id INT AUTO_INCREMENT PRIMARY KEY,
			name VARCHAR(50) NOT NULL
		)
	`)
	require.NoError(t, err)

	// 测试查询监控
	t.Run("监控慢查询", func(t *testing.T) {
		// 创建监控器，设置较低的慢查询阈值
		monitor := NewMonitor(MonitorConfig{
			SlowQueryThreshold: time.Millisecond * 1, // 1ms就算慢查询
			EnableMetrics:      true,
			MetricsPrefix:      "test",
		}, logger)

		// 监控一个正常查询
		query := "INSERT INTO test_monitor (name) VALUES (?)"
		args := []interface{}{"test"}

		// 包装查询执行
		err := monitor.WrapQuery(context.Background(), query, args, func() error {
			// 故意延迟让查询变慢
			time.Sleep(time.Millisecond * 10)
			_, err := connector.GetDB().Exec(query, args...)
			return err
		})

		assert.NoError(t, err)

		// 检查数据是否插入成功
		var count int
		err = connector.GetDB().QueryRow("SELECT COUNT(*) FROM test_monitor").Scan(&count)
		assert.NoError(t, err)
		assert.Equal(t, 1, count)
	})
}

// 测试迁移功能
func TestMigration(t *testing.T) {
	// 跳过测试如果没有Docker MySQL实例
	if testing.Short() {
		t.Skip("跳过集成测试，需要MySQL实例")
	}

	setupTestDatabase(t)
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	// 创建迁移文件
	migrationsDir := setupMigrationFiles(t)
	defer os.RemoveAll(migrationsDir)

	// 创建迁移器
	migration := NewMigration(MigrationConfig{
		MigrationsPath: migrationsDir,
		Timeout:        time.Second * 30,
		TargetVersion:  0, // 迁移到最新版本
	}, logger)

	// 执行迁移
	t.Run("执行数据库迁移", func(t *testing.T) {
		err := migration.Run(getTestDSN())
		assert.NoError(t, err)

		// 验证迁移结果
		db, err := sql.Open("mysql", getTestDSN())
		require.NoError(t, err)
		defer db.Close()

		// 检查users表是否创建
		var tableExists bool
		err = db.QueryRow("SELECT 1 FROM information_schema.tables WHERE table_schema = ? AND table_name = ?",
			testDBName, "users").Scan(&tableExists)
		if err == sql.ErrNoRows {
			tableExists = false
		} else {
			require.NoError(t, err)
		}
		assert.True(t, tableExists, "users表应该已经创建")

		// 插入测试数据
		_, err = db.Exec("INSERT INTO users (username, email) VALUES (?, ?)", "testuser", "test@example.com")
		assert.NoError(t, err)

		var username string
		err = db.QueryRow("SELECT username FROM users WHERE email = ?", "test@example.com").Scan(&username)
		assert.NoError(t, err)
		assert.Equal(t, "testuser", username)
	})
}

// 完整集成测试
func TestIntegration(t *testing.T) {
	// 跳过测试如果没有Docker MySQL实例
	if testing.Short() {
		t.Skip("跳过集成测试，需要MySQL实例")
	}

	setupTestDatabase(t)
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	// 创建迁移文件
	migrationsDir := setupMigrationFiles(t)
	defer os.RemoveAll(migrationsDir)

	// 1. 执行迁移
	migration := NewMigration(MigrationConfig{
		MigrationsPath: migrationsDir,
		Timeout:        time.Second * 30,
	}, logger)

	err := migration.Run(getTestDSN())
	require.NoError(t, err)

	// 2. 创建连接
	connector, err := NewConnector(ConnectorConfig{
		DSN:             getTestDSN(),
		MaxOpenConns:    10,
		MaxIdleConns:    5,
		ConnMaxLifetime: time.Minute * 10,
		ConnTimeout:     time.Second * 5,
	}, logger)
	require.NoError(t, err)
	defer connector.Close()

	// 3. 创建监控器
	monitor := NewMonitor(MonitorConfig{
		SlowQueryThreshold: time.Millisecond * 10,
	}, logger)

	// 4. 执行业务操作 - 插入用户并关联帖子
	t.Run("执行业务操作", func(t *testing.T) {
		// 插入用户
		var userID int64
		userQuery := "INSERT INTO users (username, email) VALUES (?, ?)"
		userArgs := []interface{}{"integration_user", "integration@example.com"}

		err := monitor.WrapQuery(context.Background(), userQuery, userArgs, func() error {
			result, err := connector.GetDB().Exec(userQuery, userArgs...)
			if err != nil {
				return err
			}
			userID, err = result.LastInsertId()
			return err
		})
		require.NoError(t, err)
		assert.Greater(t, userID, int64(0))

		// 插入帖子
		postQuery := "INSERT INTO posts (title, content, user_id) VALUES (?, ?, ?)"
		postArgs := []interface{}{"集成测试", "这是一个集成测试帖子内容", userID}

		err = monitor.WrapQuery(context.Background(), postQuery, postArgs, func() error {
			_, err := connector.GetDB().Exec(postQuery, postArgs...)
			return err
		})
		require.NoError(t, err)

		// 验证数据 - 联表查询
		var username, title string
		joinQuery := `
			SELECT u.username, p.title 
			FROM users u 
			JOIN posts p ON u.id = p.user_id 
			WHERE u.email = ?
		`
		err = connector.GetDB().QueryRow(joinQuery, "integration@example.com").Scan(&username, &title)
		assert.NoError(t, err)
		assert.Equal(t, "integration_user", username)
		assert.Equal(t, "集成测试", title)
	})
}
