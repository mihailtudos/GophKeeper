package config

import (
	"fmt"
	"sync"
	"time"

	"github.com/spf13/viper"
)

var defaultConfigPath = "./config"

const (
	AccessTokenExpiry  = 15 * time.Minute
	RefreshTokenExpiry = 7 * 24 * time.Minute // 7 days
)

type Config struct {
	HTTPServer HTTPServer   `mapstructure:"http_server"`
	Logger     LoggerConfig `mapstructure:"logger"`
	Database   Database     `mapstructure:"db"`
	Auth       Auth         `mapstructure:"auth"`
}

type HTTPServer struct {
	Port string `json:"port" yaml:"port"`
	Host string `json:"host" yaml:"host"`
}

type LoggerConfig struct {
	Level      string `json:"level" yaml:"level"`
	OutputPath string `json:"output" yaml:"output"`
	Format     string `json:"format" yaml:"format"`
}

type Database struct {
	Port     string `json:"port" yaml:"port"`
	Host     string `json:"host" yaml:"host"`
	User     string `json:"user" yaml:"user"`
	Password string `json:"password" yaml:"password"`
	DBName   string `json:"db_name" yaml:"db_name"`
	SSLMode  string `json:"ssl_mode" yaml:"ssl_mode"`
	DSN      string `json:"dsn" yaml:"dsn"`
}

type Auth struct {
	SecretKey          string        `json:"jwt_secret" yaml:"jwt_secret" mapstructure:"jwt_secret"`
	AccessTokenExpiry  time.Duration `json:"access_token_expiry" yaml:"access_token_expiry"`
	RefreshTokenExpiry time.Duration `json:"refresh_token_expiry" yaml:"refresh_token_expiry"`
}

var (
	instance *Config
	once     sync.Once
)

func MustNewConfig() *Config {
	once.Do(func() {
		// TODO: load config from env vars or file
		cfg := &Config{}

		if err := readConfig(); err != nil {
			panic(err)
		}

		if err := viper.Unmarshal(cfg); err != nil {
			panic(err)
		}

		if viper.IsSet("auth.access_token_expiry") {
			duration, err := time.ParseDuration(viper.GetString("auth.access_token_expiry"))
			if err == nil {
				cfg.Auth.AccessTokenExpiry = duration
			}
		}

		if viper.IsSet("auth.refresh_token_expiry") {
			duration, err := time.ParseDuration(viper.GetString("auth.refresh_token_expiry"))
			if err == nil {
				cfg.Auth.RefreshTokenExpiry = duration
			}
		}

		if cfg.Auth.AccessTokenExpiry == 0 {
			cfg.Auth.AccessTokenExpiry = AccessTokenExpiry // use default
		}

		if cfg.Auth.RefreshTokenExpiry == 0 {
			cfg.Auth.RefreshTokenExpiry = RefreshTokenExpiry // use default
		}

		cfg.Database.DSN = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
			cfg.Database.User,
			cfg.Database.Password,
			cfg.Database.Host,
			cfg.Database.Port,
			cfg.Database.DBName,
			cfg.Database.SSLMode)

		instance = cfg
	})

	return instance
}

func readConfig() error {
	viper.SetConfigName("app")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(defaultConfigPath) // TODO: change to flag

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			return fmt.Errorf("config file not found %w", err)
		}

		return fmt.Errorf("failed to read config file %w", err)
	}

	return nil
}
