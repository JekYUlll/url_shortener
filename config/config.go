package config

import (
	"fmt"
	"time"
)

type Config struct {
	Database  DatabaseConfig  `mapstructure:"database"`
	Redis     RedisConfig     `mapstructure:"redis"`
	Server    ServerConfig    `mapstructure:"server"`
	App       AppConfig       `mapstructure:"app"`
	ShortCode ShortCodeConfig `mapstructure:"shortcode"`
}

type DatabaseConfig struct {
	Driver       string `mapstructure:"driver"`
	Host         string `mapstructure:"host"`
	Port         int    `mapstructure:"port"`
	User         string `mapstructure:"user"`
	Password     string `mapstructure:"password"`
	DBName       string `mapstructure:"dbname"`
	SSLMode      string `mapstructure:"ssl_mode"`
	MaxIdleConns int    `mapstructure:"max_idle_conns"`
	MaxOpenConns int    `mapstructure:"max_open_conns"`
}

func (d DatabaseConfig) DNS() string {
	return fmt.Sprintf("%s://%s:%s@%s:%d/%s?sslmode=%s", d.Driver, d.User, d.Password, d.Host, d.Port, d.DBName, d.SSLMode)
}

type RedisConfig struct {
	Address  string `mapstructure:"address"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

type ServerConfig struct {
	Addr         string        `mapstructure:"addr"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
}

type AppConfig struct {
	BaseURL         string        `mapstructure:"base_url"`
	DefaultDuration time.Duration `mapstructure:"default_duration"`
	CleanupInterval time.Duration `mapstructure:"cleanup_interval"`
}

type ShortCodeConfig struct {
	Length int `mapstructure:"length"`
}
