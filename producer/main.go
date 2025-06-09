// package main

// import (
// 	"log"
// 	"producer/handlers"
// 	"producer/utils"

// 	"github.com/gofiber/fiber/v2"
// 	"github.com/gofiber/fiber/v2/middleware/cors"
// 	"github.com/gofiber/fiber/v2/middleware/logger"
// 	"github.com/gofiber/fiber/v2/middleware/recover"
// )

// func main() {
// 	cfg := utils.LoadConfig()

// 	app := fiber.New(fiber.Config{
// 		ErrorHandler: func(c *fiber.Ctx, err error) error {
// 			code := fiber.StatusInternalServerError
// 			if e, ok := err.(*fiber.Error); ok {
// 				code = e.Code
// 			}
// 			return c.Status(code).JSON(fiber.Map{
// 				"error": err.Error(),
// 			})
// 		},
// 	})

// 	// Middleware
// 	app.Use(logger.New())
// 	app.Use(recover.New())
// 	app.Use(cors.New())

// 	// Health check endpoint
// 	app.Get("/health", func(c *fiber.Ctx) error {
// 		return c.JSON(fiber.Map{"status": "healthy"})
// 	})

// 	// API routes
// 	app.Get("/ads", handlers.GetAds)
// 	app.Post("/ads/click", handlers.HandleAdClick)

// 	log.Printf("Producer service running on port %s", cfg.ProducerPort)
// 	log.Fatal(app.Listen(":" + cfg.ProducerPort))
// }

package main

import (
	"log"
	"producer/handlers"
	"producer/utils"

	// "video-ads-backend/producer/handlers"
	// "video-ads-backend/producer/utils"

	"github.com/gofiber/fiber/v2"
)

func main() {
    cfg := utils.LoadConfig()
    app := fiber.New()

    // app.Get("/ads", handlers.GetAds)
    app.Post("/ads/click", handlers.HandleAdClick)

    log.Printf("Producer service running on port %s", cfg.ProducerPort)
    log.Fatal(app.Listen(":" + cfg.ProducerPort))
}
