package main

import (
	"log"

	"github.com/RZ-ru/Inshakerov_bot/internal/config"
	"github.com/RZ-ru/Inshakerov_bot/internal/db"
	"github.com/RZ-ru/Inshakerov_bot/internal/scraper"
)

func main() {
	// 1️⃣ Загружаем конфигурацию (.env)
	cfg := config.Load()

	// 2️⃣ Подключаемся к PostgreSQL
	database := db.Connect(cfg.DBUrl)
	if database == nil {
		log.Fatal("❌ Ошибка: соединение с базой не установлено")
	}

	defer database.Close()
	log.Println("📡 Подключение к базе установлено")

	// 3️⃣ Адрес страницы с коктейлями
	url := "https://ru.inshaker.com/cocktails"

	// 4️⃣ Парсим рецепты с сайта
	log.Println("🔍 Начинаем парсинг сайта...")
	cocktails, err := scraper.ParseRecipes(url)
	if err != nil {
		log.Fatalf("Ошибка при парсинге: %v", err)
	}
	log.Printf("🍸 Найдено рецептов: %d", len(cocktails))

	// 5️⃣ Сохраняем в базу данных
	if err := db.SaveRecipes(database, cocktails); err != nil {
		log.Fatalf("Ошибка сохранения рецептов: %v", err)
	}
	log.Printf("🧾 Попытка сохранить %d рецептов...", len(cocktails))
	if err := db.SaveRecipes(database, cocktails); err != nil {
		log.Fatalf("Ошибка сохранения рецептов: %v", err)
	}
	log.Println("✅ Все рецепты успешно сохранены в базу!")
}
