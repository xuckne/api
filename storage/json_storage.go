package storage

import (
	"encoding/json"
	"myapp/models"
	"os"
	"strings"
	"sync"
)

type JSONStorage struct {
	filename string
	mu       sync.RWMutex
	data     models.LibraryStorage
}

func NewJSONStorage(filename string) (*JSONStorage, error) {
	s := &JSONStorage{
		filename: filename,
		data: models.LibraryStorage{
			Books:  []models.Book{},
			Series: []models.BookSeries{},
		},
	}

	if err := s.load(); err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	return s, nil
}

func (s *JSONStorage) load() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, err := os.Stat(s.filename); os.IsNotExist(err) {
		return nil
	}

	data, err := os.ReadFile(s.filename)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, &s.data)
}

func (s *JSONStorage) save() error {
	data, err := json.MarshalIndent(s.data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.filename, data, 0644)
}

// Утилиты поиска

func (s *JSONStorage) findBookIndex(author, title string) int {
	for i, book := range s.data.Books {
		if strings.EqualFold(book.Author, author) && strings.EqualFold(book.Title, title) {
			return i
		}
	}
	return -1
}

func (s *JSONStorage) findSeriesIndex(name string) int {
	for i, series := range s.data.Series {
		if strings.EqualFold(series.Name, name) {
			return i
		}
	}
	return -1
}

func (s *JSONStorage) updateSeries(seriesName string, book models.Book) {
	idx := s.findSeriesIndex(seriesName)
	if idx != -1 {
		// Проверка на дубликат
		for _, b := range s.data.Series[idx].Books {
			if strings.EqualFold(b.Title, book.Title) && strings.EqualFold(b.Author, book.Author) {
				return
			}
		}
		s.data.Series[idx].Books = append(s.data.Series[idx].Books, book)
		s.data.Series[idx].Total++
		if book.Read {
			s.data.Series[idx].Read++
		}
		return
	}

	newSeries := models.BookSeries{
		Name:   seriesName,
		Books:  []models.Book{book},
		Total:  1,
		Read:   0,
	}
	if book.Read {
		newSeries.Read = 1
	}
	s.data.Series = append(s.data.Series, newSeries)
}

func (s *JSONStorage) removeBookFromSeries(seriesName, title, author string) {
	sIdx := s.findSeriesIndex(seriesName)
	if sIdx == -1 {
		return
	}
	for j, book := range s.data.Series[sIdx].Books {
		if strings.EqualFold(book.Title, title) && strings.EqualFold(book.Author, author) {
			s.data.Series[sIdx].Books = append(s.data.Series[sIdx].Books[:j], s.data.Series[sIdx].Books[j+1:]...)
			s.data.Series[sIdx].Total--
			if book.Read {
				s.data.Series[sIdx].Read--
			}
			if len(s.data.Series[sIdx].Books) == 0 {
				s.data.Series = append(s.data.Series[:sIdx], s.data.Series[sIdx+1:]...)
			}
			return
		}
	}
}

func (s *JSONStorage) updateSeriesReadStatus(seriesName, title, author string, read bool) {
	sIdx := s.findSeriesIndex(seriesName)
	if sIdx == -1 {
		return
	}
	for j, book := range s.data.Series[sIdx].Books {
		if strings.EqualFold(book.Title, title) && strings.EqualFold(book.Author, author) {
			if read && !book.Read {
				s.data.Series[sIdx].Read++
			} else if !read && book.Read {
				s.data.Series[sIdx].Read--
			}
			s.data.Series[sIdx].Books[j].Read = read
			return
		}
	}
}

// === Публичные методы ===

func (s *JSONStorage) GetAllBooks() []models.Book {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.data.Books
}

func (s *JSONStorage) GetBook(author, title string) (*models.Book, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	idx := s.findBookIndex(author, title)
	if idx == -1 {
		return nil, false
	}
	book := s.data.Books[idx]
	return &book, true
}

func (s *JSONStorage) AddBook(book models.Book) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if book.Author == "" || book.Title == "" {
		return nil // валидация в handler
	}

	s.data.Books = append(s.data.Books, book)
	if book.Series != "" {
		s.updateSeries(book.Series, book)
	}
	return s.save()
}

func (s *JSONStorage) EditBook(oldAuthor, oldTitle string, newBook models.Book) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	idx := s.findBookIndex(oldAuthor, oldTitle)
	if idx == -1 {
		return nil // не найдено — handler обработает
	}

	oldBook := s.data.Books[idx]
	if newBook.Series != "" && oldBook.Series != newBook.Series {
		s.removeBookFromSeries(oldBook.Series, oldBook.Title, oldBook.Author)
		s.updateSeries(newBook.Series, newBook)
	} else if newBook.Series == "" && oldBook.Series != "" {
		s.removeBookFromSeries(oldBook.Series, oldBook.Title, oldBook.Author)
	}

	// Обновляем поля
	if newBook.Author != "" {
		s.data.Books[idx].Author = newBook.Author
	}
	if newBook.Title != "" {
		s.data.Books[idx].Title = newBook.Title
	}
	if newBook.Genre != "" {
		s.data.Books[idx].Genre = newBook.Genre
	}
	if newBook.Year >= 0 {
		s.data.Books[idx].Year = newBook.Year
	}
	if newBook.Series != "" {
		s.data.Books[idx].Series = newBook.Series
		s.data.Books[idx].SeriesOrder = newBook.SeriesOrder
	} else {
		s.data.Books[idx].Series = ""
		s.data.Books[idx].SeriesOrder = 0
	}
	// Quotes и Read не меняются через edit

	return s.save()
}

func (s *JSONStorage) RemoveBook(author, title string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	idx := s.findBookIndex(author, title)
	if idx == -1 {
		return nil
	}

	book := s.data.Books[idx]
	if book.Series != "" {
		s.removeBookFromSeries(book.Series, book.Title, book.Author)
	}

	s.data.Books = append(s.data.Books[:idx], s.data.Books[idx+1:]...)
	return s.save()
}

func (s *JSONStorage) AddQuote(author, title, quote string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	idx := s.findBookIndex(author, title)
	if idx == -1 {
		return nil
	}

	s.data.Books[idx].Quotes = append(s.data.Books[idx].Quotes, quote)
	return s.save()
}

func (s *JSONStorage) GetQuotes(author, title string) ([]string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	idx := s.findBookIndex(author, title)
	if idx == -1 {
		return nil, false
	}
	return s.data.Books[idx].Quotes, true
}

func (s *JSONStorage) MarkAsRead(author, title string, read bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	idx := s.findBookIndex(author, title)
	if idx == -1 {
		return nil
	}

	s.data.Books[idx].Read = read
	if s.data.Books[idx].Series != "" {
		s.updateSeriesReadStatus(s.data.Books[idx].Series, title, author, read)
	}
	return s.save()
}

func (s *JSONStorage) GetStatistics() models.Statistics {
	s.mu.RLock()
	defer s.mu.RUnlock()

	stats := models.Statistics{
		BooksByGenre:  make(map[string]int),
		BooksByAuthor: make(map[string]int),
		SeriesStats:   []models.SeriesStat{},
	}

	stats.TotalBooks = len(s.data.Books)
	stats.TotalSeries = len(s.data.Series)

	for _, book := range s.data.Books {
		if book.Read {
			stats.ReadBooks++
		}
		stats.BooksByGenre[book.Genre]++
		stats.BooksByAuthor[book.Author]++
	}

	if stats.TotalBooks > 0 {
		stats.ReadPercentage = float64(stats.ReadBooks) / float64(stats.TotalBooks) * 100
	}

	for _, series := range s.data.Series {
		stat := models.SeriesStat{
			Name:       series.Name,
			TotalBooks: series.Total,
			ReadBooks:  series.Read,
		}
		if series.Total > 0 {
			stat.Percentage = float64(series.Read) / float64(series.Total) * 100
		}
		stats.SeriesStats = append(stats.SeriesStats, stat)
	}

	return stats
}

func (s *JSONStorage) GetSeries() []models.BookSeries {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.data.Series
}