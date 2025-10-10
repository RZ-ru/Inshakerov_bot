package scraper

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/RZ-ru/Inshakerov_bot/internal/db"
)

// ParseRecipes — основная функция парсинга
func ParseRecipes(baseURL string) ([]db.Cocktail, error) {
	log.Println("🔍 Начинаем парсинг рецептов с сайта:", baseURL)

	resp, err := http.Get(baseURL)
	if err != nil {
		return nil, fmt.Errorf("ошибка запроса к сайту: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("ошибка HTTP: %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения HTML: %v", err)
	}

	var cocktails []db.Cocktail

	// Парсим список ссылок на коктейли
	doc.Find("a.cocktail-item-preview").Each(func(i int, s *goquery.Selection) {
		if i >= 5 { // ограничимся 5 для теста
			return
		}

		name := strings.TrimSpace(s.Find(".cocktail-item-name").Text())
		href, _ := s.Attr("href")
		img, _ := s.Find("img.cocktail-item-image").Attr("src")

		if href == "" || name == "" {
			return
		}

		cocktailURL := "https://ru.inshaker.com" + href
		imageURL := "https://ru.inshaker.com" + img

		c := db.Cocktail{
			Name:     name,
			URL:      cocktailURL,
			ImageURL: imageURL,
		}

		// Парсим страницу конкретного рецепта
		fullCocktail, err := parseCocktailDetails(c)
		if err != nil {
			log.Printf("⚠️ Ошибка парсинга %s: %v", c.Name, err)
			return
		}

		cocktails = append(cocktails, fullCocktail)
	})

	log.Printf("✅ Успешно собрали %d рецептов", len(cocktails))
	return cocktails, nil
}

// parseCocktailDetails — парсит ингредиенты и описание рецепта
func parseCocktailDetails(c db.Cocktail) (db.Cocktail, error) {
	resp, err := http.Get(c.URL)
	if err != nil {
		return c, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return c, fmt.Errorf("ошибка запроса (%d) для %s", resp.StatusCode, c.URL)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return c, err
	}

	// Извлекаем ингредиенты
	doc.Find("dl.ingredients dd.good").Each(func(_ int, s *goquery.Selection) {
		name := strings.TrimSpace(s.Find(".common-good-info").Contents().First().Text())
		amount := strings.TrimSpace(s.Find("amount").Text())
		unit := strings.TrimSpace(s.Find("unit").Text())

		if name == "" {
			return
		}

		c.Ingredients = append(c.Ingredients, db.CocktailIngredient{
			Good: db.Good{
				Name: name,
			},
			Amount: amount,
			Unit:   unit,
		})
	})

	// Извлекаем инструкцию приготовления
	c.Instructions = strings.TrimSpace(doc.Find(".how-to-make").Text())

	return c, nil
}
