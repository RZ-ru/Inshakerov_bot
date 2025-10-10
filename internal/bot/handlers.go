package bot

import (
	"database/sql"
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func HandleUpdates(bot *tgbotapi.BotAPI, update tgbotapi.Update, database *sql.DB) {
	if update.Message != nil { // обычные текстовые команды
		handleMessage(bot, update.Message, database)
	} else if update.CallbackQuery != nil { // нажатие кнопок
		handleCallback(bot, update.CallbackQuery, database)
	}
}

func handleMessage(bot *tgbotapi.BotAPI, msg *tgbotapi.Message, database *sql.DB) {
	switch msg.Text {
	case "/start":
		bot.Send(tgbotapi.NewMessage(msg.Chat.ID, "Привет! Выбери категорию: 🍸 или 🍹"))
		bot.Send(tgbotapi.NewMessage(msg.Chat.ID, "Главное меню").SetReplyMarkup(MainMenu()))
	case "🍸 Алкогольные":
		showAlcoholBases(bot, msg.Chat.ID)
	case "🍹 Безалкогольные":
		sendRecipe(bot, msg.Chat.ID, database, "non-alcoholic")
	case "⭐ Избранное":
		sendFavorites(bot, msg.Chat.ID, database, msg.From.ID)
	default:
		bot.Send(tgbotapi.NewMessage(msg.Chat.ID, "Не понимаю 😅"))
	}
}

func showAlcoholBases(bot *tgbotapi.BotAPI, chatID int64) {
	buttons := [][]tgbotapi.KeyboardButton{
		{
			tgbotapi.NewKeyboardButton("🥃 Виски"),
			tgbotapi.NewKeyboardButton("🍸 Джин"),
			tgbotapi.NewKeyboardButton("🍹 Ром"),
		},
		{
			tgbotapi.NewKeyboardButton("🍸 Водка"),
			tgbotapi.NewKeyboardButton("⬅ Назад"),
		},
	}

	reply := tgbotapi.NewReplyKeyboard(buttons...)
	reply.ResizeKeyboard = true
	bot.Send(tgbotapi.NewMessage(chatID, "Выбери базу:").SetReplyMarkup(reply))
}

func sendRecipe(bot *tgbotapi.BotAPI, chatID int64, database *sql.DB, category string) {
	recipe, err := db.GetRandomRecipe(database, category)
	if err != nil {
		bot.Send(tgbotapi.NewMessage(chatID, "Не удалось найти рецепт 😔"))
		return
	}

	msg := tgbotapi.NewPhoto(chatID, tgbotapi.FileURL(recipe.ImageURL))
	msg.Caption = formatRecipe(recipe)
	msg.ReplyMarkup = RecipeInlineKeyboard()
	bot.Send(msg)
}

func formatRecipe(r db.Recipe) string {
	return fmt.Sprintf("🍹 *%s*\n\n🧾 Ингредиенты:\n%s\n⚖️ %s\n\n👨‍🍳 Рецепт:\n%s\n\n🔗 [Оригинал](%s)",
		r.Title, r.Ingredients, r.Quantities, r.Description, r.URL)
}
