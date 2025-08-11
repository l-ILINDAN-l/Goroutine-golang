package config

import (
	"github.com/spf13/viper"
	"log"
	"os"
)

// ServerConfig holds the configuration for the HTTP server
type ServerConfig struct {
	Port string `mapstructure:"port"`
}

// PostgresConfig holds the configuration for the PostgreSQL database
type PostgresConfig struct {
	Events string `mapstructure:"events"`
}

// RedisConfig holds the configuration for the Redis cache
type RedisConfig struct {
	Addr string `mapstructure:"addr"`
}

// Config is the main configuration structure for the application, holding all sub-configurations
type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Postgres PostgresConfig `mapstructure:"postgres"`
	Redis    RedisConfig    `mapstructure:"redis"`
}

// LoadConfig reads configuration from a file and environment variables
func LoadConfig() (*Config, error) {
	v := viper.New()

	v.AutomaticEnv()
	v.AddConfigPath("./configs")
	v.SetConfigName("config")

	if err := v.ReadInConfig(); err != nil {
		log.Fatalf("error reading config file, %s", err)
		return nil, err
	}

	var config Config
	if err := v.Unmarshal(&config); err != nil {
		log.Fatalf("error unmarshalling config, %s", err)
		return nil, err
	}

	dsnTemplate := v.GetString("postgres.events")
	config.Postgres.Events = os.ExpandEnv(dsnTemplate)

	log.Printf("DEBUG: POSTGRES_USER is '%s'", os.Getenv("POSTGRES_USER"))
	log.Printf("DEBUG: POSTGRES_PASSWORD is '%s'", os.Getenv("POSTGRES_PASSWORD"))
	log.Printf("DEBUG: POSTGERS_DSN: %s", config.Postgres.Events)

	return &config, nil
}
