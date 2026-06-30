package main

import (
	"context"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"

	"github.com/MaximovIlya/vk_practice_project/internal/repository"
	"github.com/MaximovIlya/vk_practice_project/pkg/config"
)

func main() {
	cfg := config.Load()

	store, err := repository.NewStore(context.Background(), cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("db connect: %v", err)
	}
	defer store.Close()

	_ = store // передадим в хендлеры позже

	app := fiber.New(fiber.Config{
		AppName: "Pulse Quiz API",
	})

	app.Use(recover.New())
	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Authorization",
		AllowMethods: "GET, POST, PATCH, DELETE",
	}))

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	api := app.Group("/api")
	v1 := api.Group("/v1")
	_ = v1 // роуты подключим позже

	log.Fatal(app.Listen(":" + cfg.Port))
}
