package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	BotToken string
	DBUrl    string
}

func Load() *Config {
	_ = godotenv.Load() // загружаем .env, если есть
	cfg := &Config{
		BotToken: os.Getenv("BOT_TOKEN"),
		DBUrl:    os.Getenv("DB_URL"),
	}

	if cfg.BotToken == "" || cfg.DBUrl == "" {
		log.Fatal("BOT_TOKEN or DB_URL not set in environment")
	}

	return cfg
}
