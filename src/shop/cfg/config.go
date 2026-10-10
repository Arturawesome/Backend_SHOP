package cfg

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Shop     ShopConfig     `yaml:"shop"`
	Auth     AuthConfig     `yaml:"auth"`
	RabbitMq RabbitMQConfig `yaml:"rabbitmq"`
	Redis    RedisConfig    `yaml:"redis"`
}

type ShopConfig struct {
	HTTPAddr string         `yaml:"http_addr"`
	Database DatabaseConfig `yaml:"database"`
}

type AuthConfig struct {
	GRPC      GRPCConfig     `yaml:"grpc_addr"`
	Database  DatabaseConfig `yaml:"database"`
	JWTSecret string         `yaml:"jwt_secret"`
}

type GRPCConfig struct {
	ListenAddr string `yaml:"listen_addr"`
	ClientAddr string `yaml:"client_addr"`
}

type DatabaseConfig struct {
	URL string `yaml:"url"`
}

type RabbitMQConfig struct {
	URL   string `yaml:"url"`
	Queue string `yaml:"queue"`
}

type RedisConfig struct {
	Addr     string `yaml:"addr"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var conf Config
	err = yaml.Unmarshal(data, &conf)
	if err != nil {
		return nil, err
	}
	return &conf, nil
}
