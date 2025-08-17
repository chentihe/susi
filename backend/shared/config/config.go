package config

import (
	"github.com/spf13/viper"
)

type GlobalConfig struct {
	Database    DatabaseConfig
	Kafka       KafkaConfig
	Consul      ConsulConfig
	Service     ServiceConfig
	JWT         JWTConfig
	Gateway     GatewayConfig
	Development DevelopmentConfig
}

type DatabaseConfig struct {
	Host     string `mapstructure:"DB_HOST"`
	Port     int    `mapstructure:"DB_PORT"`
	Name     string `mapstructure:"DB_NAME"`
	User     string `mapstructure:"DB_USER"`
	Password string `mapstructure:"DB_PASSWORD"`
}

type KafkaConfig struct {
	Brokers string `mapstructure:"KAFKA_BROKERS"`
}

type ConsulConfig struct {
	Host string `mapstructure:"CONSUL_HOST"`
	Port int    `mapstructure:"CONSUL_PORT"`
}

type ServiceConfig struct {
	AuthService       MicroService
	PropertyService   MicroService
	TenantService     MicroService
	RenovationService MicroService
}

type MicroService struct {
	Name string
	Port int
}

type JWTConfig struct {
	SecretKey string
}

type GatewayConfig struct {
	GatewayService MicroService
}

type DevelopmentConfig struct {
	Stage string
}

func LoadConfig() (*viper.Viper, error) {
	v := viper.New()

	v.SetConfigName(".env")
	v.SetConfigType("env")
	v.AddConfigPath(".")
	v.AddConfigPath("/etc/app")
	v.AddConfigPath("/app")

	v.SetDefault("GATEWAY_PORT", 8080)
	v.SetDefault("GATEWAY_NAME", "api-gateway")

	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}

	return v, nil
}
