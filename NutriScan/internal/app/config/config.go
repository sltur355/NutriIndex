// C:\RIP\LAB1\internal\app\config\config.go
package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type Config struct {
	ServiceHost string
	ServicePort int

	Redis RedisConfig
	Minio MinioConfig
}

type RedisConfig struct {
	Host        string
	Password    string
	Port        int
	User        string
	DialTimeout time.Duration
	ReadTimeout time.Duration
}

type MinioConfig struct {
	Endpoint  string
	Bucket    string
	AccessKey string
	SecretKey string
	UseSSL    bool
}

const (
	envRedisHost = "REDIS_HOST"
	envRedisPort = "REDIS_PORT"
	envRedisUser = "REDIS_USER"
	envRedisPass = "REDIS_PASSWORD"
)

// C:\RIP\LAB1\internal\app\config\config.go

func NewConfig() (*Config, error) {
	var err error

	configName := "config"
	_ = godotenv.Load()
	if os.Getenv("CONFIG_NAME") != "" {
		configName = os.Getenv("CONFIG_NAME")
	}

	viper.SetConfigName(configName)
	viper.SetConfigType("toml")
	viper.AddConfigPath("config")
	viper.AddConfigPath(".")
	viper.WatchConfig()

	err = viper.ReadInConfig()
	if err != nil {
		return nil, err
	}
	log.Infof("USING CONFIG FILE: %s", viper.ConfigFileUsed())

	cfg := &Config{}
	err = viper.Unmarshal(cfg)
	if err != nil {
		return nil, err
	}

	// Принудительно устанавливаем порт сервиса если он не задан
	if cfg.ServicePort == 0 {
		cfg.ServicePort = 8081
		log.Warn("ServicePort not set in config, using default: 8081")
	}

	// Принудительно устанавливаем хост сервиса если он не задан
	if cfg.ServiceHost == "" {
		cfg.ServiceHost = "localhost"
	}

	// Redis config from env
	cfg.Redis.Host = os.Getenv(envRedisHost)
	if cfg.Redis.Host == "" {
		cfg.Redis.Host = "localhost"
	}

	portStr := os.Getenv(envRedisPort)
	if portStr != "" {
		cfg.Redis.Port, err = strconv.Atoi(portStr)
		if err != nil {
			return nil, fmt.Errorf("redis port must be int value: %w", err)
		}
	} else {
		cfg.Redis.Port = 6379 // порт по умолчанию
	}

	cfg.Redis.Password = os.Getenv(envRedisPass)
	cfg.Redis.User = os.Getenv(envRedisUser)

	// Устанавливаем таймауты по умолчанию если не заданы
	if cfg.Redis.DialTimeout == 0 {
		cfg.Redis.DialTimeout = 10 * time.Second
	}
	if cfg.Redis.ReadTimeout == 0 {
		cfg.Redis.ReadTimeout = 10 * time.Second
	}
	// MinIO config from env
	cfg.Minio.Endpoint = os.Getenv("MINIO_ENDPOINT")
	if cfg.Minio.Endpoint == "" {
		cfg.Minio.Endpoint = "localhost:9000" // Значение по умолчанию
	}
	cfg.Minio.AccessKey = os.Getenv("MINIO_ACCESS_KEY")
	if cfg.Minio.AccessKey == "" {
		cfg.Minio.AccessKey = "minio" // Значение по умолчанию
	}
	cfg.Minio.SecretKey = os.Getenv("MINIO_SECRET_KEY")
	if cfg.Minio.SecretKey == "" {
		cfg.Minio.SecretKey = "minio124" // Значение по умолчанию
	}
	cfg.Minio.Bucket = os.Getenv("MINIO_BUCKET")
	if cfg.Minio.Bucket == "" {
		cfg.Minio.Bucket = "biomarkers" // Значение по умолчанию
	}
	useSSLStr := os.Getenv("MINIO_USE_SSL")
	if useSSLStr != "" {
		cfg.Minio.UseSSL, _ = strconv.ParseBool(useSSLStr)
	}
	log.Infof("DEBUG: Connecting to MinIO with AccessKey=[%s] and SecretKey=[%s]", cfg.Minio.AccessKey, cfg.Minio.SecretKey)
	log.Info("config parsed")
	log.Infof("Service configuration: Host=%s, Port=%d", cfg.ServiceHost, cfg.ServicePort)

	return cfg, nil
}
