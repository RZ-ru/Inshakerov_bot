package bot

import (
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/RZ-ru/Inshakerov_bot/internal/db"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func HandleStart(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	msg := tgbotapi.NewMessage(update.Message.Chat.ID,
		"👋 Привет! Я помогу подобрать коктейль.\n\nНапиши, какой ингредиент хочешь использовать 🍋🥃")
	msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
	bot.Send(msg)
}

func HandleIngredientInput(bot *tgbotapi.BotAPI, update tgbotapi.Update, database *sql.DB) {
	text := strings.TrimSpace(strings.ToLower(update.Message.Text))
	userID := update.Message.From.ID

	var count int
	err := database.QueryRow(`SELECT COUNT(*) FROM goods WHERE LOWER(name) = $1`, text).Scan(&count)
	if err != nil {
		log.Println("Ошибка при поиске ингредиента:", err)
		bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "❌ Ошибка при обращении к базе. Попробуй позже."))
		return
	}

	if count == 0 {
		rows, err := database.Query(`
	SELECT name FROM goods 
	WHERE LOWER(name) % $1 OR LOWER(name) ILIKE '%' || $1 || '%' 
	LIMIT 3;
`, text)
		if err != nil {
			log.Println("Ошибка поиска похожих ингредиентов:", err)
			bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "⚠️ Ошибка поиска похожих ингредиентов."))
			return
		}
		defer rows.Close()

		var suggestions []string
		for rows.Next() {
			var name string
			rows.Scan(&name)
			suggestions = append(suggestions, name)
		}

		if len(suggestions) > 0 {
			similar := suggestions[0]
			msg := tgbotapi.NewMessage(update.Message.Chat.ID,
				fmt.Sprintf("🤔 Возможно, вы имели в виду *%s*?", similar))
			msg.ParseMode = "Markdown"
			msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
				tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData("Да", "confirm_"+similar),
					tgbotapi.NewInlineKeyboardButtonData("Нет", "reject"),
				),
			)
			bot.Send(msg)
			return
		}

		bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "🥲 Такой ингредиент не найден. Попробуй другой."))
		return
	}

	ShowCocktailsForIngredient(bot, update.Message.Chat.ID, database, text, userID)
}

func HandleIngredientConfirm(bot *tgbotapi.BotAPI, update tgbotapi.Update, database *sql.DB) {
	data := update.CallbackQuery.Data
	userID := update.CallbackQuery.From.ID
	chatID := update.CallbackQuery.Message.Chat.ID

	if strings.HasPrefix(data, "confirm_") {
		ingredient := strings.TrimPrefix(data, "confirm_")
		bot.Request(tgbotapi.NewCallback(update.CallbackQuery.ID, ""))
		bot.Send(tgbotapi.NewMessage(chatID, fmt.Sprintf("🔍 Ищу рецепты с ингредиентом: %s...", ingredient)))
		ShowCocktailsForIngredient(bot, chatID, database, ingredient, userID)
	} else if data == "reject" {
		msg := tgbotapi.NewMessage(chatID, "Окей 🙂 напиши ингредиент ещё раз:")
		bot.Send(msg)
	}
}

func ShowCocktailsForIngredient(bot *tgbotapi.BotAPI, chatID int64, database *sql.DB, ingredient string, userID int64) {
	cocktails, err := db.GetCocktailsBySimilarIngredients(database, ingredient)
	if err != nil {
		bot.Send(tgbotapi.NewMessage(chatID, "❌ Ошибка при поиске рецептов."))
		return
	}

	if len(cocktails) == 0 {
		bot.Send(tgbotapi.NewMessage(chatID, "🥲 Коктейлей с таким ингредиентом не найдено."))
		return
	}

	msg := tgbotapi.NewMessage(chatID,
		fmt.Sprintf("🍸 Найдено %d рецептов с ингредиентом *%s*!", len(cocktails), ingredient))
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = IngredientMenuKeyboard()
	bot.Send(msg)
}

func HandleCallback(bot *tgbotapi.BotAPI, update tgbotapi.Update, database *sql.DB) {
	data := update.CallbackQuery.Data
	userID := update.CallbackQuery.From.ID
	chatID := update.CallbackQuery.Message.Chat.ID

	// Ответим Telegram, чтобы убрать "часики" с кнопки
	bot.Request(tgbotapi.NewCallback(update.CallbackQuery.ID, ""))

	// Разбираем действие, например fav_123
	parts := strings.Split(data, "_")
	if len(parts) < 2 {
		return
	}

	action := parts[0]
	cocktailIDStr := parts[1]
	cocktailID, err := strconv.Atoi(cocktailIDStr)
	if err != nil {
		log.Println("Ошибка парсинга callback ID:", err)
		return
	}

	switch action {
	case "fav":
		err := db.AddFavorite(database, int64(userID), cocktailID)
		if err != nil {
			send(bot, chatID, "❌ Не удалось добавить в избранное.")
			return
		}
		send(bot, chatID, "💛 Добавлено в избранное!")

	case "ignore":
		err := db.AddIgnored(database, int64(userID), cocktailID)
		if err != nil {
			send(bot, chatID, "❌ Ошибка при скрытии коктейля.")
			return
		}
		send(bot, chatID, "🚫 Коктейль скрыт.")

	case "next":
		// Заглушка — позже добавим выбор следующего
		send(bot, chatID, "⏭ Показать следующий пока не реализовано 🙂")
	default:
		log.Printf("Неизвестное действие: %s", data)
	}
}

func send(bot *tgbotapi.BotAPI, chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	bot.Send(msg)
}
