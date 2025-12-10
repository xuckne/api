package main

import (
	"myapp/handlers"
	"myapp/storage"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

func main() {
	// Инициализация хранилища
	storage, err := storage.NewJSONStorage("library.json")
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}

	// Инициализация обработчиков
	bookHandler := handlers.NewBookHandler(storage)

	// Создание Fiber приложения
	app := fiber.New(fiber.Config{
		AppName: "My Library API",
	})

	// Middleware
	app.Use(logger.New())
	app.Use(recover.New())

	// Маршруты
	api := app.Group("/api/v1")

	// Книги
	books := api.Group("/books")
	{
		books.Get("/", bookHandler.GetAllBooks)
		books.Get("/:author/:title", bookHandler.GetBook) // author + title как идентификатор
		books.Post("/", bookHandler.AddBook)
		books.Put("/:author/:title", bookHandler.EditBook)
		books.Delete("/:author/:title", bookHandler.RemoveBook)
	}

	// Цитаты
	quotes := api.Group("/quotes")
	{
		quotes.Post("/", bookHandler.AddQuote)
		quotes.Get("/:author/:title", bookHandler.GetQuotes)
	}

	// Отметка прочитанного
	read := api.Group("/read")
	{
		read.Post("/:author/:title", bookHandler.MarkAsRead)
	}

	// Статистика
	api.Get("/stats", bookHandler.GetStatistics)

	// Серии
	api.Get("/series", bookHandler.GetSeries)

	// Корень
	app.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message": "My Library API is running",
			"version": "1.0.0",
		})
	})

	// 404
	app.Use(func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Endpoint not found",
		})
	})

	log.Println("Server starting on :3000")
	log.Fatal(app.Listen(":3000"))
}