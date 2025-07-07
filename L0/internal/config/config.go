package config

import (
	"github.com/spf13/viper"
	"log"
	"os"
)

type PostgresShardConfig struct {
	Primary  string   `mapstructure:"primary"`
	Replicas []string `mapstructure:"replicas"`
}

type PostgresConfig struct {
	LookupDbDsn string                         `mapstructure:"lookup_db_dsn"`
	Shards      map[string]PostgresShardConfig `mapstructure:"shards"`
}

type RedisConfig struct {
	Addr string `mapstructure:"addr"`
}

type KafkaConfig struct {
	Brokers []string `mapstructure:"brokers"`
	Topic   string   `mapstructure:"topic"`
}

type ServerConfig struct {
	Port string `mapstructure:"port"`
}

type Config struct {
	Postgres PostgresConfig `mapstructure:"postgres"`
	Redis    RedisConfig    `mapstructure:"redis"`
	Kafka    KafkaConfig    `mapstructure:"kafka"`
	Server   ServerConfig   `mapstructure:"server"`
}

func MustLoad() *Config {
	v := viper.New()

	v.AutomaticEnv()
	v.AddConfigPath("./configs")
	v.SetConfigName("config")

	if err := v.ReadInConfig(); err != nil {
		log.Fatalf("error reading config file: %v", err)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		log.Fatalf("error unmarshaling config: %v", err)
	}

	log.Printf("DEBUG: POSTGRES_USER is '%s'", os.Getenv("POSTGRES_USER"))
	log.Printf("DEBUG: POSTGRES_PASSWORD is '%s'", os.Getenv("POSTGRES_PASSWORD"))

	cfg.Postgres.LookupDbDsn = os.ExpandEnv(cfg.Postgres.LookupDbDsn)
	for key, shard := range cfg.Postgres.Shards {
		shard.Primary = os.ExpandEnv(shard.Primary)

		for i, replica := range shard.Replicas {
			shard.Replicas[i] = os.ExpandEnv(replica)
		}

		cfg.Postgres.Shards[key] = shard
	}

	return &cfg
}
