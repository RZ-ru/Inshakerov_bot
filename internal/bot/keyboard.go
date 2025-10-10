package bot

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

// Главное меню
func MainMenu() tgbotapi.ReplyKeyboardMarkup {
	keyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("🍸 Алкогольные"),
			tgbotapi.NewKeyboardButton("🍹 Безалкогольные"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("⭐ Избранное"),
			tgbotapi.NewKeyboardButton("➕ Добавить ингредиент"),
			tgbotapi.NewKeyboardButton("🔄 Сбросить"),
		),
	)
	keyboard.ResizeKeyboard = true
	return keyboard
}

// Инлайн-кнопки под рецептом
func RecipeInlineKeyboard() tgbotapi.InlineKeyboardMarkup {
	buttons := [][]tgbotapi.InlineKeyboardButton{
		{
			tgbotapi.NewInlineKeyboardButtonData("⭐ В избранное", "add_fav"),
			tgbotapi.NewInlineKeyboardButtonData("🔁 Следующий", "next_recipe"),
		},
	}
	return tgbotapi.NewInlineKeyboardMarkup(buttons...)
}
