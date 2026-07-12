package config

import (
	"log"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Env          string `mapstructure:"ENV"`
	Port         string `mapstructure:"PORT"`
	DBHost       string `mapstructure:"DB_HOST"`
	DBPort       string `mapstructure:"DB_PORT"`
	DBUser       string `mapstructure:"DB_USER"`
	DBPassword   string `mapstructure:"DB_PASSWORD"`
	DBName       string `mapstructure:"DB_NAME"`
	DBSslMode    string `mapstructure:"DB_SSLMODE"`
	RedisHost    string `mapstructure:"REDIS_HOST"`
	RedisPort    string `mapstructure:"REDIS_PORT"`
	RedisPassword string `mapstructure:"REDIS_PASSWORD"`
	RedisDB      int    `mapstructure:"REDIS_DB"`
	RedisURL     string `mapstructure:"REDIS_URL"`
	KafkaBrokers string `mapstructure:"KAFKA_BROKERS"` // comma-separated
	KafkaTopic   string `mapstructure:"KAFKA_TOPIC"`
	KafkaGroupID string `mapstructure:"KAFKA_GROUP_ID"`
	JWTSecret    string `mapstructure:"JWT_SECRET"`
	JWTAccessTTL time.Duration `mapstructure:"JWT_ACCESS_TTL"`
	JWTRefreshTTL time.Duration `mapstructure:"JWT_REFRESH_TTL"`
	AdminEmail   string `mapstructure:"ADMIN_EMAIL"`
	AdminPassword string `mapstructure:"ADMIN_PASSWORD"`
	LemonSqueezyWebhookSecret string `mapstructure:"LEMON_SQUEEZY_WEBHOOK_SECRET"`
	LemonSqueezyProVariantID  string `mapstructure:"LEMON_SQUEEZY_PRO_VARIANT_ID"`
}

func LoadConfig(path string) (*Config, error) {
	viper.AddConfigPath(path)
	viper.SetConfigName(".env")
	viper.SetConfigType("env")

	// Read environment variables if present
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Set defaults
	viper.SetDefault("ENV", "development")
	viper.SetDefault("PORT", "8080")
	viper.SetDefault("DB_HOST", "localhost")
	viper.SetDefault("DB_PORT", "5432")
	viper.SetDefault("DB_USER", "postgres")
	viper.SetDefault("DB_PASSWORD", "postgres")
	viper.SetDefault("DB_NAME", "ratelimiter")
	viper.SetDefault("DB_SSLMODE", "disable")
	viper.SetDefault("REDIS_HOST", "localhost")
	viper.SetDefault("REDIS_PORT", "6379")
	viper.SetDefault("REDIS_PASSWORD", "")
	viper.SetDefault("REDIS_DB", 0)
	viper.SetDefault("REDIS_URL", "")
	viper.SetDefault("KAFKA_BROKERS", "localhost:9092")
	viper.SetDefault("KAFKA_TOPIC", "api_logs")
	viper.SetDefault("KAFKA_GROUP_ID", "analytics_consumers")
	viper.SetDefault("JWT_SECRET", "super_secret_jwt_key_please_change_in_production")
	viper.SetDefault("JWT_ACCESS_TTL", 15*time.Minute)
	viper.SetDefault("JWT_REFRESH_TTL", 7*24*time.Hour)
	viper.SetDefault("ADMIN_EMAIL", "admin@ratelimiter.io")
	viper.SetDefault("ADMIN_PASSWORD", "AdminPass123!")
	viper.SetDefault("LEMON_SQUEEZY_WEBHOOK_SECRET", "")
	viper.SetDefault("LEMON_SQUEEZY_PRO_VARIANT_ID", "")

	if err := viper.ReadInConfig(); err != nil {
		// It's ok if .env file is missing, config can be read from environment variables directly
		log.Printf("Warning: .env file not found or could not be read: %v. Using environment variables.", err)
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, nil
}
