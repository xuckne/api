package handlers

import (
	"myapp/models"
	"myapp/storage"
	"strings"

	"github.com/gofiber/fiber/v2"
)

type BookHandler struct {
	store *storage.JSONStorage
}

func NewBookHandler(store *storage.JSONStorage) *BookHandler {
	return &BookHandler{store: store}
}

// Запросы

type AddBookRequest struct {
	Author      string `json:"author" validate:"required"`
	Title       string `json:"title" validate:"required"`
	Genre       string `json:"genre" validate:"required"`
	Year        int    `json:"year" validate:"min=0,max=9999"`
	Series      string `json:"series,omitempty"`
	SeriesOrder int    `json:"series_order,omitempty"`
}

type EditBookRequest struct {
	Author      string `json:"author,omitempty"`
	Title       string `json:"title,omitempty"`
	Genre       string `json:"genre,omitempty"`
	Year        int    `json:"year"`
	Series      string `json:"series,omitempty"`
	SeriesOrder int    `json:"series_order,omitempty"`
}

type MarkReadRequest struct {
	Read bool `json:"read"`
}

// === Endpoints ===

func (h *BookHandler) GetAllBooks(c *fiber.Ctx) error {
	books := h.store.GetAllBooks()
	return c.JSON(fiber.Map{"books": books, "count": len(books)})
}

func (h *BookHandler) GetBook(c *fiber.Ctx) error {
	author := strings.TrimSpace(c.Params("author"))
	title := strings.TrimSpace(c.Params("title"))

	book, found := h.store.GetBook(author, title)
	if !found {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Book not found"})
	}
	return c.JSON(book)
}

func (h *BookHandler) AddBook(c *fiber.Ctx) error {
	var req AddBookRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid JSON"})
	}

	if req.Author == "" || req.Title == "" || req.Genre == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Author, title, and genre are required"})
	}
	if req.Year < 0 || req.Year > 9999 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid year"})
	}

	book := models.Book{
		Author:      strings.TrimSpace(req.Author),
		Title:       strings.TrimSpace(req.Title),
		Genre:       strings.TrimSpace(req.Genre),
		Year:        req.Year,
		Read:        false,
		Quotes:      []string{},
		Series:      strings.TrimSpace(req.Series),
		SeriesOrder: req.SeriesOrder,
	}

	if err := h.store.AddBook(book); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to add book"})
	}

	return c.Status(fiber.StatusCreated).JSON(book)
}

func (h *BookHandler) EditBook(c *fiber.Ctx) error {
	oldAuthor := strings.TrimSpace(c.Params("author"))
	oldTitle := strings.TrimSpace(c.Params("title"))

	var req EditBookRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid JSON"})
	}

	_, found := h.store.GetBook(oldAuthor, oldTitle)
	if !found {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Book not found"})
	}

	newBook := models.Book{
		Author:      strings.TrimSpace(req.Author),
		Title:       strings.TrimSpace(req.Title),
		Genre:       strings.TrimSpace(req.Genre),
		Year:        req.Year,
		Series:      strings.TrimSpace(req.Series),
		SeriesOrder: req.SeriesOrder,
	}

	if err := h.store.EditBook(oldAuthor, oldTitle, newBook); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update book"})
	}

	updated, _ := h.store.GetBook(newBook.Author, newBook.Title)
	return c.JSON(updated)
}

func (h *BookHandler) RemoveBook(c *fiber.Ctx) error {
	author := strings.TrimSpace(c.Params("author"))
	title := strings.TrimSpace(c.Params("title"))

	_, found := h.store.GetBook(author, title)
	if !found {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Book not found"})
	}

	if err := h.store.RemoveBook(author, title); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to delete book"})
	}

	return c.SendStatus(fiber.StatusNoContent)
}

func (h *BookHandler) AddQuote(c *fiber.Ctx) error {
	author := strings.TrimSpace(c.Params("author"))
	title := strings.TrimSpace(c.Params("title"))

	type QuoteReq struct {
		Quote string `json:"quote" validate:"required"`
	}
	var req QuoteReq
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid JSON"})
	}

	if req.Quote == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Quote is required"})
	}

	if err := h.store.AddQuote(author, title, strings.TrimSpace(req.Quote)); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to add quote"})
	}

	return c.JSON(fiber.Map{"message": "Quote added"})
}

func (h *BookHandler) GetQuotes(c *fiber.Ctx) error {
	author := strings.TrimSpace(c.Params("author"))
	title := strings.TrimSpace(c.Params("title"))

	quotes, found := h.store.GetQuotes(author, title)
	if !found {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Book not found"})
	}

	return c.JSON(fiber.Map{"quotes": quotes})
}

func (h *BookHandler) MarkAsRead(c *fiber.Ctx) error {
	author := strings.TrimSpace(c.Params("author"))
	title := strings.TrimSpace(c.Params("title"))

	var req MarkReadRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid JSON"})
	}

	if err := h.store.MarkAsRead(author, title, req.Read); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update read status"})
	}

	return c.JSON(fiber.Map{"message": "Read status updated"})
}

func (h *BookHandler) GetStatistics(c *fiber.Ctx) error {
	stats := h.store.GetStatistics()
	return c.JSON(stats)
}

func (h *BookHandler) GetSeries(c *fiber.Ctx) error {
	series := h.store.GetSeries()
	return c.JSON(series)
}