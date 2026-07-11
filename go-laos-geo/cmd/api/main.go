package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
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
	app.Use(logger.New())

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

	log.Println("Server is running on :3000")
	log.Fatal(app.Listen(":3000"))
}
