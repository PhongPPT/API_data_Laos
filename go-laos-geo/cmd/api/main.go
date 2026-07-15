package main

import (
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"github.com/user/go-laos-geo/config"
	"github.com/user/go-laos-geo/delivery/http"
	"github.com/user/go-laos-geo/repository/postgres"
	"github.com/user/go-laos-geo/usecase"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	db, err := sqlx.Connect("postgres", cfg.GetDSN())
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	geoRepo := postgres.NewGeoRepository(db)
	geoUsecase := usecase.NewGeoUsecase(geoRepo)

	app := fiber.New()

	// Add CORS middleware to allow frontend requests
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept",
	}))

	app.Use(logger.New())

	// Add Built-in Rate Limiting (replaces Kong for free tier)
	app.Use(limiter.New(limiter.Config{
		Max:        100,             // Maximum 100 requests
		Expiration: 1 * time.Minute, // Per 1 minute
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP() // Limit by IP address
		},
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error": "Too Many Requests. Please wait a minute.",
			})
		},
	}))

	http.NewGeoHandler(app, geoUsecase)

	// Serve the openapi JSON for Scalar
	app.Static("/openapi.json", "./docs/openapi.json")

	// Serve Scalar API Docs using a simple HTML string
	app.Get("/reference", func(c *fiber.Ctx) error {
		html := `<!doctype html>
<html>
  <head>
    <title>Laos Geo API Docs</title>
    <meta charset="utf-8" />
    <meta
      name="viewport"
      content="width=device-width, initial-scale=1" />
  </head>
  <body>
    <script id="api-reference" data-url="/openapi.json"></script>
    <script src="https://cdn.jsdelivr.net/npm/@scalar/api-reference"></script>
  </body>
</html>`
		c.Set("Content-Type", "text/html")
		return c.SendString(html)
	})

	log.Println("Server is running on :3005")
	log.Fatal(app.Listen(":3005"))
}
