package handlers

import (
	"producer/models"

	"github.com/gofiber/fiber/v2"
)

func GetAds(c *fiber.Ctx) error {
	// In a real application, fetch ads from a database
	// For now, return mock data
	ads := []models.Ad{
		{
			ID:        "ad1",
			ImageURL:  "https://example.com/ad1.png",
			TargetURL: "https://example.com/target1",
		},
		{
			ID:        "ad2", 
			ImageURL:  "https://example.com/ad2.png",
			TargetURL: "https://example.com/target2",
		},
		{
			ID:        "ad3",
			ImageURL:  "https://example.com/ad3.png", 
			TargetURL: "https://example.com/target3",
		},
	}
	
	return c.JSON(fiber.Map{
		"ads": ads,
		"total": len(ads),
	})
}