// pkg/config/loader.go
package config

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Database DatabaseConfig `yaml:"database"`
	MQ       MQConfig       `yaml:"mq"`
	Storage  StorageConfig  `yaml:"storage"`
	Server   ServerConfig   `yaml:"server"`
	Logger   LoggerConfig   `yaml:"logger"`
	Security SecurityConfig `yaml:"security"`
}

type DatabaseConfig struct {
	MySQL struct {
		Host         string `yaml:"host"`
		Port         int    `yaml:"port"`
		Username     string `yaml:"username"`
		Password     string `yaml:"password"`
		Name         string `yaml:"name"`
		MaxOpenConns int    `yaml:"max_open_conns" mapstructure:"max_open_conns"`
	} `yaml:"mysql"`
	Redis struct {
		Host     string `yaml:"host"`
		Port     int    `yaml:"port"`
		Password string `yaml:"password"`
		DB       int    `yaml:"db"`
	} `yaml:"redis"`
}

type MQConfig struct {
	RabbitMQ struct {
		Host     string `yaml:"host"`
		Port     int    `yaml:"port"`
		Username string `yaml:"username"`
		Password string `yaml:"password"`
		VHost    string `yaml:"vhost"`
		Consumer struct {
			PrefetchCount int           `yaml:"prefetch_count" mapstructure:"prefetch_count"`
			RetryInterval time.Duration `yaml:"retry_interval" mapstructure:"retry_interval"`
		} `yaml:"consumer"`
	} `yaml:"rabbitmq"`
}

type StorageConfig struct {
	MinIO struct {
		Endpoint  string `yaml:"endpoint"`
		AccessKey string `yaml:"access_key" mapstructure:"access_key"`
		SecretKey string `yaml:"secret_key" mapstructure:"secret_key"`
		UseSSL    bool   `yaml:"use_ssl" mapstructure:"use_ssl"`
		Bucket    string `yaml:"bucket"`
	} `yaml:"minio"`
}

type ServerConfig struct {
	HTTP struct {
		Port    int           `yaml:"port"`
		Timeout time.Duration `yaml:"timeout"`
	} `yaml:"http"`
	GRPC struct {
		Port    int           `yaml:"port"`
		Timeout time.Duration `yaml:"timeout"`
	} `yaml:"grpc"`
}

type LoggerConfig struct {
	Level            string   `yaml:"level"`
	Encoding         string   `yaml:"encoding"`
	OutputPaths      []string `yaml:"output_paths" mapstructure:"output_paths"`
	ErrorOutputPaths []string `yaml:"error_output_paths" mapstructure:"error_output_paths"`
}

type SecurityConfig struct {
	JWTSecret string `yaml:"jwt_secret" mapstructure:"jwt_secret"`
	AESKey    string `yaml:"aes_key" mapstructure:"aes_key"`
}

func validateConfig(cfg *Config) error {
	if cfg.Database.MySQL.Host == "" {
		return errors.New("mysql host is required")
	}

	if cfg.Database.MySQL.Password == "" && os.Getenv("DB_PASSWORD") == "" {
		return errors.New("missing database password")
	}

	if len(cfg.Security.AESKey) != 32 {
		return fmt.Errorf("aes key must be 32 bytes, got %d bytes", len(cfg.Security.AESKey))
	}

	return nil
}

func Load(env string) (*Config, error) {
	v := viper.New()

	// 基础配置
	v.SetConfigName("base")
	// 设置多级搜索路径
	v.AddConfigPath("./configs")        // 项目根目录执行时
	v.AddConfigPath("../../configs")    // test/unit目录执行时
	v.AddConfigPath("../../../configs") // test/integration目录执行时

	v.SetConfigType("yaml")

	// 读取环境配置
	if env != "" {
		v.SetConfigName(env)
		v.AddConfigPath("./configs")
	}

	// 环境变量支持
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// 合并配置
	if err := v.MergeInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	// 处理环境变量替换
	replaceEnvVariables(v)

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	if err := validateConfig(&cfg); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &cfg, nil
}

func replaceEnvVariables(v *viper.Viper) {
	for _, key := range v.AllKeys() {
		val := v.GetString(key)
		if strings.HasPrefix(val, "${") && strings.HasSuffix(val, "}") {
			envVar := strings.TrimSuffix(strings.TrimPrefix(val, "${"), "}")
			defaultValue := ""

			if parts := strings.SplitN(envVar, ":", 2); len(parts) > 1 {
				envVar = parts[0]
				defaultValue = parts[1]
			}

			if envVal := os.Getenv(envVar); envVal != "" {
				v.Set(key, envVal)
			} else if defaultValue != "" {
				v.Set(key, defaultValue)
			}
		}
	}
}

// GetRabbitMQURL 根据配置生成RabbitMQ连接URL
func (c *Config) GetRabbitMQURL() string {
	rabbitCfg := c.MQ.RabbitMQ

	// 构建URL，格式为：amqp://username:password@host:port/vhost
	vhost := rabbitCfg.VHost
	if vhost != "" && vhost[0] != '/' {
		vhost = "/" + vhost
	}

	return fmt.Sprintf("amqp://%s:%s@%s:%d%s",
		rabbitCfg.Username,
		rabbitCfg.Password,
		rabbitCfg.Host,
		rabbitCfg.Port,
		vhost)
}

// GetMySQLDSN 根据配置生成MySQL DSN
func (c *Config) GetMySQLDSN() string {
	mysqlCfg := c.Database.MySQL

	// 使用标准DSN格式
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		mysqlCfg.Username,
		mysqlCfg.Password,
		mysqlCfg.Host,
		mysqlCfg.Port,
		mysqlCfg.Name)
}
