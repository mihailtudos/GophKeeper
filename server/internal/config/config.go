package config

import (
	"fmt"
	"sync"

	"github.com/spf13/viper"
)

const (
	defaultConfigPath = "./internal/config"
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
	Level  string `json:"level" yaml:"level"`
	Output string `json:"output" yaml:"output"`
	Format string `json:"format" yaml:"format"`
}

type Database struct {
	Port     string `json:"port" yaml:"port"`
	Host     string `json:"host" yaml:"host"`
	User     string `json:"user" yaml:"user"`
	Password string `json:"password" yaml:"password"`
	DBName   string `json:"db_name" yaml:"db_name"`
	SSLMode  string `json:"ssl_mode" yaml:"ssl_mode"`
}

type Auth struct {
	SecretKey string `json:"jwt_secret" yaml:"jwt_secret"`
}

var (
	instance *Config
	once     sync.Once
)

func NewConfig() *Config {
	once.Do(func() {
		// TODO: load config from env vars or file
		cfg := &Config{}
		if err := readConfig(); err != nil {
			panic(err)
		}

		if err := viper.Unmarshal(cfg); err != nil {
			panic(err)
		}

		instance = cfg
	})

	return instance
}

func readConfig() error {
	viper.SetConfigName("server")
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
