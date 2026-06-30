package main

import (
	"context"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	"github.com/golang-migrate/migrate/v4/source/iofs"

	"github.com/MaximovIlya/vk_practice_project/internal/auth"
	"github.com/MaximovIlya/vk_practice_project/internal/handler"
	"github.com/MaximovIlya/vk_practice_project/internal/middleware"
	"github.com/MaximovIlya/vk_practice_project/internal/repository"
	"github.com/MaximovIlya/vk_practice_project/internal/service"
	"github.com/MaximovIlya/vk_practice_project/internal/ws"
	"github.com/MaximovIlya/vk_practice_project/pkg/config"
	db "github.com/MaximovIlya/vk_practice_project/db/sqlc"
	"github.com/MaximovIlya/vk_practice_project/db/migrations"
)

func main() {
	cfg := config.Load()

	// Run migrations before opening the connection pool
	if err := runMigrations(cfg.DatabaseURL); err != nil {
		log.Fatalf("migrate: %v", err)
	}

	store, err := repository.NewStore(context.Background(), cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("db connect: %v", err)
	}
	defer store.Close()

	jwtManager, err := auth.NewManager(cfg.JWTSecret, cfg.JWTAccessTokenTTL, cfg.JWTRefreshTokenTTL)
	if err != nil {
		log.Fatalf("jwt manager: %v", err)
	}

	
	authSvc := service.NewAuthService(store, jwtManager)
	quizSvc := service.NewQuizService(store)
	sessSvc := service.NewSessionService(store)

	
	authH := handler.NewAuthHandler(authSvc)
	quizH := handler.NewQuizHandler(quizSvc)
	sessH := handler.NewSessionHandler(sessSvc)
	playH := handler.NewPlayHandler(store)

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

	api := app.Group("/api/v1")

	
	authR := api.Group("/auth")
	authR.Post("/register", authH.Register)
	authR.Post("/login", authH.Login)
	authR.Post("/refresh", authH.Refresh)
	authR.Post("/forgot-password", authH.ForgotPassword)
	authR.Post("/reset-password", authH.ResetPassword)

	
	authR.Get("/me", middleware.Auth(jwtManager), authH.Me)
	authR.Post("/select-role", middleware.Auth(jwtManager), authH.SelectRole)

	
	quizR := api.Group("/quiz", middleware.Auth(jwtManager), middleware.RequireRole(db.RoleORGANIZER))
	quizR.Get("/", quizH.ListMine)
	quizR.Post("/", quizH.Create)
	quizR.Get("/:id", quizH.GetByID)
	quizR.Patch("/:id", quizH.Update)
	quizR.Delete("/:id", quizH.Delete)
	quizR.Post("/:id/questions", quizH.CreateQuestion)
	quizR.Patch("/:id/questions/:questionId", quizH.UpdateQuestion)
	quizR.Delete("/:id/questions/:questionId", quizH.DeleteQuestion)
	quizR.Post("/:id/session", sessH.GetOrCreate)
	quizR.Get("/:id/session", sessH.Get)
	quizR.Delete("/:id/session", sessH.Delete)
	quizR.Get("/:id/history", playH.GetQuizHistory)


	api.Get("/play/:code", middleware.Auth(jwtManager), playH.JoinByCode)
	api.Get("/results/:sessionId", middleware.Auth(jwtManager), playH.GetResults)

	
	hub := ws.NewHub(store)
	go hub.Run()
	app.Get("/ws", ws.Upgrade, ws.Handler(hub, jwtManager))

	log.Fatal(app.Listen(":" + cfg.Port))
}

func runMigrations(databaseURL string) error {
	d, err := iofs.New(migrations.FS, ".")
	if err != nil {
		return err
	}
	// golang-migrate pgx/v5 driver expects pgx5:// scheme
	pgxURL := "pgx5://" + databaseURL[len("postgresql://"):]
	m, err := migrate.NewWithSourceInstance("iofs", d, pgxURL)
	if err != nil {
		return err
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}
	log.Println("migrations applied")
	return nil
}
