// pkg/config/loader.go
package config

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/blackarbiter/go-sac/pkg/domain"
	"github.com/blackarbiter/go-sac/pkg/scanner"
	"github.com/spf13/viper"
)

type Config struct {
	Database DatabaseConfig `yaml:"database"`
	MQ       MQConfig       `yaml:"mq"`
	Storage  StorageConfig  `yaml:"storage"`
	Server   ServerConfig   `yaml:"server"`
	Logger   LoggerConfig   `yaml:"logger"`
	Security SecurityConfig `yaml:"security"`
	Scanner  ScannerConfig  `yaml:"scanner"`
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

type ScannerConfig struct {
	Concurrency struct {
		MaxWorkers int `yaml:"max_workers" mapstructure:"max_workers"`
		QueueSize  int `yaml:"queue_size" mapstructure:"queue_size"`
	} `yaml:"concurrency" mapstructure:"concurrency"`

	// 统一的熔断器配置
	CircuitBreaker struct {
		Threshold         uint32        `yaml:"threshold" mapstructure:"threshold"`                   // 总错误阈值
		CriticalThreshold uint32        `yaml:"critical_threshold" mapstructure:"critical_threshold"` // 严重错误阈值
		ResetTimeout      time.Duration `yaml:"reset_timeout" mapstructure:"reset_timeout"`           // 重置超时时间
	} `yaml:"circuit_breaker" mapstructure:"circuit_breaker"`

	SAST struct {
		ResourceProfile struct {
			MinCPU   int `yaml:"min_cpu" mapstructure:"min_cpu"`
			MaxCPU   int `yaml:"max_cpu" mapstructure:"max_cpu"`
			MemoryMB int `yaml:"memory_mb" mapstructure:"memory_mb"`
		} `yaml:"resource_profile" mapstructure:"resource_profile"`
		SecurityProfile struct {
			RunAsUser  int  `yaml:"run_as_user" mapstructure:"run_as_user"`
			RunAsGroup int  `yaml:"run_as_group" mapstructure:"run_as_group"`
			NoNewPrivs bool `yaml:"no_new_privs" mapstructure:"no_new_privs"`
		} `yaml:"security_profile" mapstructure:"security_profile"`
		Timeout time.Duration `yaml:"timeout" mapstructure:"timeout"`
	} `yaml:"sast" mapstructure:"sast"`
	DAST struct {
		ResourceProfile struct {
			MinCPU   int `yaml:"min_cpu" mapstructure:"min_cpu"`
			MaxCPU   int `yaml:"max_cpu" mapstructure:"max_cpu"`
			MemoryMB int `yaml:"memory_mb" mapstructure:"memory_mb"`
		} `yaml:"resource_profile" mapstructure:"resource_profile"`
		SecurityProfile struct {
			RunAsUser  int  `yaml:"run_as_user" mapstructure:"run_as_user"`
			RunAsGroup int  `yaml:"run_as_group" mapstructure:"run_as_group"`
			NoNewPrivs bool `yaml:"no_new_privs" mapstructure:"no_new_privs"`
		} `yaml:"security_profile" mapstructure:"security_profile"`
		Timeout time.Duration `yaml:"timeout" mapstructure:"timeout"`
	} `yaml:"dast" mapstructure:"dast"`
	SCA struct {
		ResourceProfile struct {
			MinCPU   int `yaml:"min_cpu" mapstructure:"min_cpu"`
			MaxCPU   int `yaml:"max_cpu" mapstructure:"max_cpu"`
			MemoryMB int `yaml:"memory_mb" mapstructure:"memory_mb"`
		} `yaml:"resource_profile" mapstructure:"resource_profile"`
		SecurityProfile struct {
			RunAsUser  int  `yaml:"run_as_user" mapstructure:"run_as_user"`
			RunAsGroup int  `yaml:"run_as_group" mapstructure:"run_as_group"`
			NoNewPrivs bool `yaml:"no_new_privs" mapstructure:"no_new_privs"`
		} `yaml:"security_profile" mapstructure:"security_profile"`
		Timeout time.Duration `yaml:"timeout" mapstructure:"timeout"`
	} `yaml:"sca" mapstructure:"sca"`
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

func Load(server string) (*Config, error) {
	v := viper.New()
	v.SetConfigType("yaml")

	// 1. 强制先加载 base.yaml
	v.SetConfigName("base")
	v.AddConfigPath("./configs")
	v.AddConfigPath("../../configs")
	v.AddConfigPath("../../../configs")
	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to load base config: %w", err)
	}

	// 2. 加载环境特定配置（如 scan.yaml）
	if server != "" {
		v.SetConfigName(server)
		// 注意：此处不需要重复添加 ConfigPath，Viper 会复用之前的路径配置
		if err := v.MergeInConfig(); err != nil {
			return nil, fmt.Errorf("failed to merge %s config: %w", server, err)
		}
	}

	// 3. 处理环境变量替换
	replaceEnvVariables(v)

	// 4. 解析配置到结构体
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// 5. 验证配置
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

func (c *Config) GetTaskApiBaseURL() string {
	return fmt.Sprintf("http://%s:%d", "127.0.0.1", 8088)
}

func (c *Config) GetAuthToken() string {
	return "1234567890"
}

// GetRedisAddr 获取Redis地址
func (c *Config) GetRedisAddr() string {
	return fmt.Sprintf("%s:%d", c.Database.Redis.Host, c.Database.Redis.Port)
}

// GetRedisPassword 获取Redis密码
func (c *Config) GetRedisPassword() string {
	return c.Database.Redis.Password
}

// GetRedisDB 获取Redis数据库编号
func (c *Config) GetRedisDB() int {
	return c.Database.Redis.DB
}

// GetRedisPoolSize 获取Redis连接池大小
func (c *Config) GetRedisPoolSize() int {
	return 10 // 默认连接池大小
}

// GetScannerConfig 获取扫描器配置
func (c *Config) GetScannerConfig(scanType domain.ScanType) (scanner.ResourceProfile, scanner.SecurityConfig, time.Duration) {
	switch scanType {
	case domain.ScanTypeStaticCodeAnalysis:
		return scanner.ResourceProfile{
				MinCPU:   c.Scanner.SAST.ResourceProfile.MinCPU,
				MaxCPU:   c.Scanner.SAST.ResourceProfile.MaxCPU,
				MemoryMB: c.Scanner.SAST.ResourceProfile.MemoryMB,
			}, scanner.SecurityConfig{
				RunAsUser:                int64(c.Scanner.SAST.SecurityProfile.RunAsUser),
				RunAsGroup:               int64(c.Scanner.SAST.SecurityProfile.RunAsGroup),
				AllowPrivilegeEscalation: !c.Scanner.SAST.SecurityProfile.NoNewPrivs,
			}, c.Scanner.SAST.Timeout
	case domain.ScanTypeDast:
		return scanner.ResourceProfile{
				MinCPU:   c.Scanner.DAST.ResourceProfile.MinCPU,
				MaxCPU:   c.Scanner.DAST.ResourceProfile.MaxCPU,
				MemoryMB: c.Scanner.DAST.ResourceProfile.MemoryMB,
			}, scanner.SecurityConfig{
				RunAsUser:                int64(c.Scanner.DAST.SecurityProfile.RunAsUser),
				RunAsGroup:               int64(c.Scanner.DAST.SecurityProfile.RunAsGroup),
				AllowPrivilegeEscalation: !c.Scanner.DAST.SecurityProfile.NoNewPrivs,
			}, c.Scanner.DAST.Timeout
	case domain.ScanTypeSca:
		return scanner.ResourceProfile{
				MinCPU:   c.Scanner.SCA.ResourceProfile.MinCPU,
				MaxCPU:   c.Scanner.SCA.ResourceProfile.MaxCPU,
				MemoryMB: c.Scanner.SCA.ResourceProfile.MemoryMB,
			}, scanner.SecurityConfig{
				RunAsUser:                int64(c.Scanner.SCA.SecurityProfile.RunAsUser),
				RunAsGroup:               int64(c.Scanner.SCA.SecurityProfile.RunAsGroup),
				AllowPrivilegeEscalation: !c.Scanner.SCA.SecurityProfile.NoNewPrivs,
			}, c.Scanner.SCA.Timeout
	default:
		return scanner.ResourceProfile{
				MinCPU:   2,
				MaxCPU:   4,
				MemoryMB: 2048,
			}, scanner.SecurityConfig{
				RunAsUser:                int64(1001),
				RunAsGroup:               int64(1001),
				AllowPrivilegeEscalation: false,
			}, 180 * time.Second
	}
}

// GetCircuitBreakerConfig 获取熔断器配置
func (c *Config) GetCircuitBreakerConfig() (uint32, uint32, time.Duration) {
	return c.Scanner.CircuitBreaker.Threshold,
		c.Scanner.CircuitBreaker.CriticalThreshold,
		c.Scanner.CircuitBreaker.ResetTimeout
}

// GetConcurrencyConfig 获取全局并行配置文件
func (c *Config) GetConcurrencyConfig() (int, int) {
	return c.Scanner.Concurrency.MaxWorkers, c.Scanner.Concurrency.QueueSize
}
