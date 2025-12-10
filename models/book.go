package models

type Book struct {
	Title       string   `json:"title"`
	Author      string   `json:"author"`
	Genre       string   `json:"genre"`
	Year        int      `json:"year"`
	Read        bool     `json:"read"`
	Quotes      []string `json:"quotes"`
	Series      string   `json:"series,omitempty"`
	SeriesOrder int      `json:"series_order,omitempty"`
}

type BookKey struct {
	Author string
	Title  string
}

type BookSeries struct {
	Name   string `json:"name"`
	Books  []Book `json:"books"`
	Total  int    `json:"total"`
	Read   int    `json:"read"`
}

type SeriesStat struct {
	Name       string  `json:"name"`
	TotalBooks int     `json:"total_books"`
	ReadBooks  int     `json:"read_books"`
	Percentage float64 `json:"percentage"`
}

type Statistics struct {
	TotalBooks     int                 `json:"total_books"`
	ReadBooks      int                 `json:"read_books"`
	ReadPercentage float64             `json:"read_percentage"`
	BooksByGenre   map[string]int      `json:"books_by_genre"`
	BooksByAuthor  map[string]int      `json:"books_by_author"`
	TotalSeries    int                 `json:"total_series"`
	SeriesStats    []SeriesStat        `json:"series_stats"`
}

// Хранилище совместимо с исходной структурой
type LibraryStorage struct {
	Books  []Book      `json:"books"`
	Series []BookSeries `json:"series"`
}