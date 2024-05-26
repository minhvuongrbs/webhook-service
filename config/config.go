package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/minhvuongrbs/webhook-service/internal/common/httpclient"
	"github.com/minhvuongrbs/webhook-service/pkg/database"
	pkgfkafka "github.com/minhvuongrbs/webhook-service/pkg/kafka"
	"github.com/minhvuongrbs/webhook-service/pkg/logging"
	pkgredis "github.com/minhvuongrbs/webhook-service/pkg/redis"
	"github.com/minhvuongrbs/webhook-service/pkg/temporal"
	"github.com/spf13/viper"
)

type Config struct {
	Env                  string                     `mapstructure:"env"`
	Logger               logging.Config             `mapstructure:"logger"`
	Database             database.Config            `mapstructure:"database"`
	RedisConnection      pkgredis.Config            `mapstructure:"redis_connection"`
	KafkaSubscriberEvent pkgfkafka.SubscriberConfig `mapstructure:"kafka_subscriber_event"`
	DeadLetterProducer   pkgfkafka.ProducerConfig   `mapstructure:"dead_letter_producer"`
	Temporal             temporal.Config            `mapstructure:"temporal"`
	HttpClient           httpclient.Config          `mapstructure:"http_client"`
}

func loadDefaultConfig() *Config {
	return &Config{
		Env:      "local",
		Logger:   logging.Config{},
		Database: database.Config{},
	}
}

func LoadConfig(configPath string) (Config, error) {
	viper.SetConfigFile(configPath)
	viper.SetConfigType("yaml")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "__"))
	viper.AutomaticEnv()
	/**
	|-------------------------------------------------------------------------
	| You should set default config value here
	| 1. Populate the default value in (Source code)
	| 2. Then merge from config (YAML) and OS environment
	|-----------------------------------------------------------------------*/
	// Load default config
	c := loadDefaultConfig()
	configBuffer, err := json.Marshal(c)
	if err != nil {
		return Config{}, fmt.Errorf("failed to marshal default config: %w", err)
	}

	//1. Populate the default value in (Source code)
	if err = viper.ReadConfig(bytes.NewBuffer(configBuffer)); err != nil {
		return Config{}, fmt.Errorf("failed to read config: %w", err)
	}

	//2. Then merge from config (YAML) and OS environment
	if err = viper.MergeInConfig(); err != nil {
		return Config{}, fmt.Errorf("failed to merge in config: %w", err)
	}
	// Populate all config again
	err = viper.Unmarshal(c)
	if err != nil {
		return Config{}, fmt.Errorf("failed to unmarshal config: %w", err)
	}
	return *c, err
}
