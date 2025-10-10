package main

import (
	"log"

	"github.com/RZ-ru/Inshakerov_bot/internal/config"
	"github.com/RZ-ru/Inshakerov_bot/internal/db"
	"github.com/RZ-ru/Inshakerov_bot/internal/scraper"
)

func main() {
	cfg := config.Load()
	database := db.Connect(cfg.DBUrl)
	defer database.Close()

	url := "https://ru.inshaker.com/cocktails"
	recipes, err := scraper.ParseRecipes(url)
	if err != nil {
		log.Fatal(err)
	}

	// Сохраняем в базу
	db.SaveRecipes(database, recipes)

	log.Println("✅ Готово. Проверь таблицу recipes в pgAdmin.")
}
