package main

import (
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/RZ-ru/Inshakerov_bot/internal/bot"
	"github.com/RZ-ru/Inshakerov_bot/internal/config"
	"github.com/RZ-ru/Inshakerov_bot/internal/db"
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

	// 3️⃣ Инициализируем Telegram-бота
	botAPI, err := tgbotapi.NewBotAPI(cfg.BotToken)
	if err != nil {
		log.Fatalf("❌ Ошибка запуска бота: %v", err)
	}
	botAPI.Debug = false
	log.Printf("🤖 Бот запущен как %s", botAPI.Self.UserName)

	// 4️⃣ Настраиваем получение обновлений
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := botAPI.GetUpdatesChan(u)

	// 5️⃣ Основной цикл
	for update := range updates {
		if update.Message != nil {
			switch {
			case update.Message.IsCommand():
				switch update.Message.Command() {
				case "start":
					bot.HandleStart(botAPI, update)
				}
			default:
				bot.HandleIngredientInput(botAPI, update, database)
			}
		} else if update.CallbackQuery != nil {
			bot.HandleIngredientConfirm(botAPI, update, database)
			bot.HandleCallback(botAPI, update, database)
		}
	}
}
