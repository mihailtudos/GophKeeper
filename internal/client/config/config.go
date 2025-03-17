package config

import (
	"flag"
	"fmt"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"sync"
)

var defaultConfigPath = "./config/client_config.yaml"

type Config struct {
	HTTPServer HTTPServer   `mapstructure:"http_server"`
	Logger     LoggerConfig `mapstructure:"logger"`
}

type HTTPServer struct {
	Port string `json:"port" yaml:"port"`
	Host string `json:"host" yaml:"host"`
}

type LoggerConfig struct {
	Level      string `mapstructure:"level" json:"level" yaml:"level"`
	OutputPath string `mapstructure:"output" json:"output" yaml:"output"`
	Format     string `mapstructure:"format" json:"format" yaml:"format"`
}

var (
	instance *Config
	once     sync.Once
)

func MustNewConfig() *Config {
	once.Do(func() {
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
	viper.SetConfigFile(configPath())

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			return fmt.Errorf("config file not found %w", err)
		}

		return fmt.Errorf("failed to read config file %w", err)
	}

	return nil
}

func configPath() string {
	var res string

	flag.StringVar(&res, "c", defaultConfigPath, "path to config file")
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()
	_ = viper.BindPFlags(pflag.CommandLine)
	res = viper.GetString("c")

	if err := viper.BindEnv("CLIENT_CONFIG_PATH"); err == nil {
		return res
	}

	var ClientConfigPath = viper.GetString("CLIENT_CONFIG_PATH")
	if ClientConfigPath != "" {
		res = ClientConfigPath
	}

	return res
}
