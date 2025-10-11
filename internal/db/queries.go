package db

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/lib/pq"
)

// SaveRecipes — сохраняет список коктейлей в базу
func SaveRecipes(db *sql.DB, cocktails []Cocktail) error {
	for _, cocktail := range cocktails {
		// 1️⃣ Добавляем коктейль
		cocktailID, err := insertCocktail(db, cocktail)
		if err != nil {
			log.Printf("❌ Ошибка добавления коктейля %s: %v", cocktail.Name, err)
			continue
		}

		// 2️⃣ Добавляем ингредиенты и связи
		for _, ing := range cocktail.Ingredients {
			goodID, err := getOrCreateGood(db, ing.Good.Name)
			if err != nil {
				log.Printf("⚠️ Ошибка при добавлении ингредиента %s: %v", ing.Good.Name, err)
				continue
			}

			err = insertCocktailIngredient(db, cocktailID, goodID, ing.Amount, ing.Unit)
			if err != nil {
				log.Printf("⚠️ Ошибка при добавлении связи %s -> %s: %v", cocktail.Name, ing.Good.Name, err)
			}
		}
	}

	log.Printf("✅ Успешно сохранено %d коктейлей", len(cocktails))
	return nil
}

// insertCocktail — вставляет коктейль и возвращает его ID
func insertCocktail(db *sql.DB, c Cocktail) (int, error) {
	var id int
	err := db.QueryRow(`
		INSERT INTO cocktails (name, url, image_url, instructions)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (name) DO UPDATE
		    SET url = EXCLUDED.url,
		        image_url = EXCLUDED.image_url,
		        instructions = EXCLUDED.instructions
		RETURNING id;
	`, c.Name, c.URL, c.ImageURL, c.Instructions).Scan(&id)

	if err == sql.ErrNoRows {
		// если обновление без RETURNING
		err = db.QueryRow(`SELECT id FROM cocktails WHERE name = $1`, c.Name).Scan(&id)
	}
	return id, err
}

// getOrCreateGood — возвращает ID ингредиента, создавая новый при необходимости
func getOrCreateGood(db *sql.DB, name string) (int, error) {
	var id int

	err := db.QueryRow(`SELECT id FROM goods WHERE name = $1`, name).Scan(&id)
	if err == sql.ErrNoRows {
		err = db.QueryRow(`
			INSERT INTO goods (name) VALUES ($1)
			ON CONFLICT (name) DO UPDATE SET name = EXCLUDED.name
			RETURNING id;
		`, name).Scan(&id)
	}
	return id, err
}

// insertCocktailIngredient — создаёт связь коктейль ↔ ингредиент
func insertCocktailIngredient(db *sql.DB, cocktailID, goodID int, amount, unit string) error {
	_, err := db.Exec(`
		INSERT INTO cocktail_ingredients (cocktail_id, good_id, amount, unit)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT DO NOTHING;
	`, cocktailID, goodID, amount, unit)
	return err
}

// GetCocktailsByIngredients — поиск коктейлей по списку ингредиентов
func GetCocktailsByIngredients(db *sql.DB, ingredients []string) ([]Cocktail, error) {
	if len(ingredients) == 0 {
		return nil, fmt.Errorf("список ингредиентов пуст")
	}

	query := `
		SELECT c.id, c.name, c.url, c.image_url, c.instructions
		FROM cocktails c
		JOIN cocktail_ingredients ci ON c.id = ci.cocktail_id
		JOIN goods g ON ci.good_id = g.id
		WHERE g.name = ANY($1)
		GROUP BY c.id
		HAVING COUNT(DISTINCT g.name) = $2;
	`

	rows, err := db.Query(query, pq.Array(ingredients), len(ingredients))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []Cocktail
	for rows.Next() {
		var c Cocktail
		if err := rows.Scan(&c.ID, &c.Name, &c.URL, &c.ImageURL, &c.Instructions); err != nil {
			return nil, err
		}
		result = append(result, c)
	}

	return result, nil
}

// AddFavorite — добавить коктейль в избранное
func AddFavorite(db *sql.DB, userID int64, cocktailID int) error {
	_, err := db.Exec(`
		INSERT INTO favorites (user_id, cocktail_id)
		VALUES ($1, $2)
		ON CONFLICT (user_id, cocktail_id) DO NOTHING;
	`, userID, cocktailID)
	return err
}

// RemoveFavorite — удалить коктейль из избранного
func RemoveFavorite(db *sql.DB, userID int64, cocktailID int) error {
	_, err := db.Exec(`
		DELETE FROM favorites
		WHERE user_id = $1 AND cocktail_id = $2;
	`, userID, cocktailID)
	return err
}

// GetFavorites — получить список избранных коктейлей пользователя
func GetFavorites(db *sql.DB, userID int64) ([]Cocktail, error) {
	rows, err := db.Query(`
		SELECT c.id, c.name, c.url, c.image_url, c.instructions
		FROM cocktails c
		JOIN favorites f ON c.id = f.cocktail_id
		WHERE f.user_id = $1
		ORDER BY f.created_at DESC;
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []Cocktail
	for rows.Next() {
		var c Cocktail
		if err := rows.Scan(&c.ID, &c.Name, &c.URL, &c.ImageURL, &c.Instructions); err != nil {
			return nil, err
		}
		result = append(result, c)
	}
	return result, nil
}

// AddIgnored — добавить коктейль в игнор
func AddIgnored(db *sql.DB, userID int64, cocktailID int) error {
	_, err := db.Exec(`
		INSERT INTO ignored (user_id, cocktail_id)
		VALUES ($1, $2)
		ON CONFLICT (user_id, cocktail_id) DO NOTHING;
	`, userID, cocktailID)
	return err
}

// RemoveIgnored — удалить коктейль из игнора
func RemoveIgnored(db *sql.DB, userID int64, cocktailID int) error {
	_, err := db.Exec(`
		DELETE FROM ignored
		WHERE user_id = $1 AND cocktail_id = $2;
	`, userID, cocktailID)
	return err
}

// GetIgnored — получить список игнорированных коктейлей пользователя
func GetIgnored(db *sql.DB, userID int64) ([]Cocktail, error) {
	rows, err := db.Query(`
		SELECT c.id, c.name, c.url, c.image_url, c.instructions
		FROM cocktails c
		JOIN ignored i ON c.id = i.cocktail_id
		WHERE i.user_id = $1
		ORDER BY i.created_at DESC;
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []Cocktail
	for rows.Next() {
		var c Cocktail
		if err := rows.Scan(&c.ID, &c.Name, &c.URL, &c.ImageURL, &c.Instructions); err != nil {
			return nil, err
		}
		result = append(result, c)
	}
	return result, nil
}

// GetCocktailsBySimilarIngredients ищет коктейли, где ингредиенты похожи по названию
func GetCocktailsBySimilarIngredients(db *sql.DB, ingredient string) ([]Cocktail, error) {
	query := `
		SELECT DISTINCT c.id, c.name, c.url, c.image_url, c.instructions
		FROM cocktails c
		JOIN cocktail_ingredients ci ON c.id = ci.cocktail_id
		JOIN goods g ON ci.good_id = g.id
		WHERE LOWER(g.name) ILIKE '%' || $1 || '%';
	`

	rows, err := db.Query(query, ingredient)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cocktails []Cocktail
	for rows.Next() {
		var c Cocktail
		if err := rows.Scan(&c.ID, &c.Name, &c.URL, &c.ImageURL, &c.Instructions); err != nil {
			return nil, err
		}
		cocktails = append(cocktails, c)
	}

	return cocktails, nil
}
