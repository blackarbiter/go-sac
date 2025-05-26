package config

import (
	"github.com/spf13/viper"
)

// Config 应用配置
type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	RabbitMQ RabbitMQConfig `mapstructure:"rabbitmq"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	HTTP HTTPConfig `mapstructure:"http"`
}

// HTTPConfig HTTP服务配置
type HTTPConfig struct {
	Port int `mapstructure:"port"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Type                   string `mapstructure:"type"`
	Host                   string `mapstructure:"host"`
	Port                   int    `mapstructure:"port"`
	Username               string `mapstructure:"username"`
	Password               string `mapstructure:"password"`
	Name                   string `mapstructure:"name"`
	MaxIdleConns           int    `mapstructure:"max_idle_conns"`
	MaxOpenConns           int    `mapstructure:"max_open_conns"`
	ConnMaxLifetimeMinutes int    `mapstructure:"conn_max_lifetime_minutes"`
}

// RabbitMQConfig RabbitMQ配置
type RabbitMQConfig struct {
	URL string `mapstructure:"url"`
}

// Load 加载配置
func Load(configPath string) (*Config, error) {
	v := viper.New()

	// 设置默认值
	setDefaults(v)

	// 如果提供了配置文件路径，则从文件加载
	if configPath != "" {
		v.SetConfigFile(configPath)
		if err := v.ReadInConfig(); err != nil {
			return nil, err
		}
	}

	// 从环境变量加载
	v.AutomaticEnv()

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// setDefaults 设置默认配置
func setDefaults(v *viper.Viper) {
	// 服务器默认配置
	v.SetDefault("server.http.port", 8080)

	// 数据库默认配置
	v.SetDefault("database.type", "sqlite")
	v.SetDefault("database.name", "task.db")
	v.SetDefault("database.max_idle_conns", 10)
	v.SetDefault("database.max_open_conns", 100)
	v.SetDefault("database.conn_max_lifetime_minutes", 60)

	// RabbitMQ默认配置
	v.SetDefault("rabbitmq.url", "amqp://guest:guest@localhost:5672/")
}
