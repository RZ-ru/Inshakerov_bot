package bot

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

// Меню после выбора ингредиента
func IngredientMenuKeyboard() tgbotapi.ReplyKeyboardMarkup {
	return tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("👀 Показать"),
			tgbotapi.NewKeyboardButton("➕ Добавить ингредиент"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("⭐ Избранное"),
			tgbotapi.NewKeyboardButton("🧹 Очистить ингредиенты"),
		),
	)
}
