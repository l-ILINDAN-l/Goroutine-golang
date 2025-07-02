package config

import (
	// "database/sql"
	"github.com/spf13/viper"
	"log"
)

type DSN string

type PostgresShardConfig struct {
	Primary  DSN   `yaml:"primary"`
	Replicas []DSN `yaml:"replicas"`
}

type PostgresConfig struct {
	LookupDbDsn DSN                            `yaml:"lookup_db_dsn"`
	Shards      map[string]PostgresShardConfig `yaml:"shards"`
}

type RedisConfig struct {
	Addr string `yaml:"addr"`
}

type KafkaConfig struct {
	Brokers []string `yaml:"brokers"`
	Topic   string   `yaml:"topic"`
}

type ServerConfig struct {
	Port string `yaml:"port"`
}

type Config struct {
	Postgres PostgresConfig `yaml:"postgres"`
	Redis    RedisConfig    `yaml:"redis"`
	Kafka    KafkaConfig    `yaml:"tkafka"`
	Server   ServerConfig   `yaml:"server"`
}

func MustLoad() *Config {
	v := viper.New()

	v.AutomaticEnv()
	v.AddConfigPath("./configs") // Указываем путь
	v.SetConfigName("config")    // Указываем имя файла без расширения

	if err := v.ReadInConfig(); err != nil {
		log.Fatalf("error reading config file: %v", err)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		log.Fatalf("error unmarshaling config: %v", err)
	}

	return &cfg
}
