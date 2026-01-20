package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

// Config 应用配置结构体
type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	App      AppConfig      `mapstructure:"app"`
	Database DatabaseConfig `mapstructure:"database"`
	JWT      JWTConfig      `mapstructure:"jwt"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Port int `mapstructure:"port"`
}

// AppConfig 应用配置
type AppConfig struct {
	Env  string `mapstructure:"env"`
	Name string `mapstructure:"name"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	DBName   string `mapstructure:"dbname"`
	SSLMode  string `mapstructure:"sslmode"`
	Timezone string `mapstructure:"timezone"`
}

// JWTConfig JWT 配置
type JWTConfig struct {
	Secret      string `mapstructure:"secret"`
	ExpireHours int    `mapstructure:"expire_hours"`
}

// DSN 返回 PostgreSQL 连接字符串
func (d *DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%d sslmode=%s TimeZone=%s",
		d.Host, d.User, d.Password, d.DBName, d.Port, d.SSLMode, d.Timezone,
	)
}

// AppConfig 全局配置实例
var AppCfg *Config

// LoadConfig 加载配置文件
func LoadConfig(path string) (*Config, error) {
	viper.SetConfigFile(path)
	viper.SetConfigType("yaml")

	// 支持环境变量覆盖 (例如 DATABASE_HOST)
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	AppCfg = &cfg
	return &cfg, nil
}
