package config

import (
// "database/sql"
// "github.com/spf13/viper"
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
	Kafka    KafkaConfig    `yaml:"kafka"`
	Server   ServerConfig   `yaml:"server"`
}

//func NewConfig() *Config {}
