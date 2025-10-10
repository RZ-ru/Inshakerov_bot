package db

// Cocktail — основная сущность: рецепт коктейля
type Cocktail struct {
	ID           int
	Name         string
	URL          string
	ImageURL     string
	Instructions string
	Ingredients  []CocktailIngredient // список связей с ингредиентами
}

// Good — справочник ингредиентов (уникальные записи)
type Good struct {
	ID       int
	Name     string
	Category string // опционально, например: "Фрукты", "Алкоголь"
	ImageURL string // опционально: картинка ингредиента
}

// CocktailIngredient — связь между коктейлем и ингредиентом (многие-ко-многим)
type CocktailIngredient struct {
	ID         int
	CocktailID int
	GoodID     int
	Good       Good   // для удобства, чтобы не делать отдельный JOIN при парсинге
	Amount     string // например "50"
	Unit       string // например "мл", "г"
}
