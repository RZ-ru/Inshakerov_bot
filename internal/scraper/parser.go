package scraper

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/RZ-ru/Inshakerov_bot/internal/db"
)

const (
	maxPages             = 60                      // максимум страниц
	pageDelay            = 1200 * time.Millisecond // пауза между страницами
	detailDelay          = 300 * time.Millisecond  // пауза между запросами к рецептам
	requestTimeout       = 20 * time.Second
	baseHost             = "https://ru.inshaker.com"
	listItemSelector     = "a.cocktail-item-preview"
	ingredientSelector   = "dl.ingredients dd.good"
	instructionsSelector = ".how-to-make"
)

var httpClient = &http.Client{Timeout: requestTimeout}

// ParseRecipes — парсит все рецепты со страниц ?random_page=
func ParseRecipes(baseURL string) ([]db.Cocktail, error) {
	log.Println("🔍 Запуск постраничного парсинга по random_page:", baseURL)

	all := make([]db.Cocktail, 0, 1200)
	seen := make(map[string]struct{})
	emptyCount := 0

	for page := 1; page <= maxPages; page++ {
		url := fmt.Sprintf("%s?random_page=%d", baseURL, page)
		log.Printf("📄 Страница %d → %s", page, url)

		doc, err := fetchDoc(url)
		if err != nil {
			log.Printf("⚠️ Ошибка загрузки страницы %d: %v", page, err)
			continue
		}

		pageCocktails := parseCocktailList(doc, seen)
		log.Printf("✅ Страница %d — собрано %d рецептов (итого: %d)", page, len(pageCocktails), len(all))

		if len(pageCocktails) == 0 {
			emptyCount++
			if emptyCount >= 2 {
				log.Println("ℹ️ Две пустые страницы подряд — завершаем обход.")
				break
			}
		} else {
			emptyCount = 0
			all = append(all, pageCocktails...)
		}

		time.Sleep(pageDelay)
	}

	log.Printf("🍸 Всего собрано рецептов: %d", len(all))
	return all, nil
}

// fetchDoc — получает и разбирает HTML
func fetchDoc(url string) (*goquery.Document, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; InshakerBot/1.0)")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	return goquery.NewDocumentFromReader(resp.Body)
}

// parseCocktailList — извлекает карточки коктейлей со страницы
func parseCocktailList(doc *goquery.Document, seen map[string]struct{}) []db.Cocktail {
	var cocktails []db.Cocktail

	doc.Find(listItemSelector).Each(func(_ int, s *goquery.Selection) {
		name := strings.TrimSpace(s.Find(".cocktail-item-name").Text())
		href, _ := s.Attr("href")
		img, _ := s.Find("img.cocktail-item-image").Attr("src")

		if name == "" || href == "" {
			return
		}

		cocktailURL := baseHost + href
		if _, ok := seen[cocktailURL]; ok {
			return
		}

		imageURL := ""
		if strings.HasPrefix(img, "/") {
			imageURL = baseHost + img
		}

		c := db.Cocktail{
			Name:     name,
			URL:      cocktailURL,
			ImageURL: imageURL,
		}

		full, err := parseCocktailDetails(c)
		if err != nil {
			log.Printf("⚠️ Ошибка деталей [%s]: %v", c.Name, err)
			return
		}

		cocktails = append(cocktails, full)
		seen[cocktailURL] = struct{}{}
		time.Sleep(detailDelay)
	})

	return cocktails
}

// parseCocktailDetails — парсит ингредиенты и инструкцию рецепта
func parseCocktailDetails(c db.Cocktail) (db.Cocktail, error) {
	doc, err := fetchDoc(c.URL)
	if err != nil {
		return c, err
	}

	doc.Find(ingredientSelector).Each(func(_ int, s *goquery.Selection) {
		name := strings.TrimSpace(s.Find(".common-good-info").Contents().First().Text())
		amount := strings.TrimSpace(s.Find("amount").Text())
		unit := strings.TrimSpace(s.Find("unit").Text())

		styleAttr, exists := s.Find(".icon").Attr("style")
		imageURL := ""
		if exists && strings.Contains(styleAttr, "background-image") {
			start := strings.Index(styleAttr, "url(")
			end := strings.Index(styleAttr, ");")
			if start != -1 && end != -1 && end > start+4 {
				path := styleAttr[start+4 : end]
				imageURL = baseHost + strings.Trim(path, "'\"")
			}
		}

		if name == "" {
			return
		}

		c.Ingredients = append(c.Ingredients, db.CocktailIngredient{
			Good: db.Good{
				Name:     name,
				ImageURL: imageURL,
			},
			Amount: amount,
			Unit:   unit,
		})
	})

	c.Instructions = strings.TrimSpace(doc.Find(instructionsSelector).Text())
	return c, nil
}
