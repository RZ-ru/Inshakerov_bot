package main

import (
	"log"

	"github.com/RZ-ru/Inshakerov_bot/internal/bot"
	"github.com/RZ-ru/Inshakerov_bot/internal/config"
	"github.com/RZ-ru/Inshakerov_bot/internal/db"
)

func main() {
	cfg := config.Load()
	database := db.Connect(cfg.DBUrl)
	defer database.Close()

	log.Println("Starting bot...")
	bot.Run(cfg.BotToken)
}
