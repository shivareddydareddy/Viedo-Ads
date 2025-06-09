// package handlers

// import (
// 	"encoding/json"
// 	"log"
// 	"producer/kafka"
// 	"producer/utils"
// 	"time"

// 	// "video-ads-backend/producer/kafka"
// 	// "video-ads-backend/producer/utils"

// 	"github.com/gofiber/fiber/v2"
// )

// type ClickEvent struct {
//     AdID            string    `json:"ad_id"`
//     Timestamp       time.Time `json:"timestamp"`
//     IP              string    `json:"ip"`
//     PlaybackSeconds int       `json:"playback_seconds"`
// }

// func HandleAdClick(c *fiber.Ctx) error {
//     var input struct {
//         AdID            string `json:"ad_id"`
//         PlaybackSeconds int    `json:"playback_seconds"`
//     }

//     if err := c.BodyParser(&input); err != nil {
//         return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request"})
//     }

//     event := ClickEvent{
//         AdID:            input.AdID,
//         Timestamp:       time.Now().UTC(),
//         IP:              c.IP(),
//         PlaybackSeconds: input.PlaybackSeconds,
//     }

//     data, err := json.Marshal(event)
//     if err != nil {
//         return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to serialize event"})
//     }

//     cfg := utils.LoadConfig()
//     if err := kafka.PublishMessage(cfg.KafkaBroker, cfg.KafkaTopic, data); err != nil {
//         log.Printf("Failed to publish message: %v", err)
//         return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to log click"})
//     }

//     return c.JSON(fiber.Map{"status": "click logged"})
// }




package handlers

import (
	"encoding/json"
	"log"
	"net"
	"producer/kafka"
	"producer/utils"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

type ClickEvent struct {
	AdID            string    `json:"ad_id"`
	Timestamp       time.Time `json:"timestamp"`
	IP              string    `json:"ip"`
	PlaybackSeconds int       `json:"playback_seconds"`
	UserAgent       string    `json:"user_agent,omitempty"`
}

type ClickRequest struct {
	AdID            string `json:"ad_id" validate:"required"`
	PlaybackSeconds int    `json:"playback_seconds" validate:"min=0"`
}

func HandleAdClick(c *fiber.Ctx) error {
	var input ClickRequest

	// Parse request body
	if err := c.BodyParser(&input); err != nil {
		log.Printf("Failed to parse request body: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request format",
		})
	}

	// Validate required fields
	if input.AdID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "ad_id is required",
		})
	}

	if input.PlaybackSeconds < 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "playback_seconds must be non-negative",
		})
	}

	// Get client IP (handle proxy headers)
	clientIP := getClientIP(c)

	// Create click event
	event := ClickEvent{
		AdID:            input.AdID,
		Timestamp:       time.Now().UTC(),
		IP:              clientIP,
		PlaybackSeconds: input.PlaybackSeconds,
		UserAgent:       c.Get("User-Agent"),
	}

	// Serialize event to JSON
	data, err := json.Marshal(event)
	if err != nil {
		log.Printf("Failed to serialize click event: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to process click event",
		})
	}

	// Load configuration
	cfg := utils.LoadConfig()

	// Publish to Kafka (non-blocking)
	go func() {
		if err := kafka.PublishMessage(cfg.KafkaBroker, cfg.KafkaTopic, data); err != nil {
			log.Printf("Failed to publish click event to Kafka: %v", err)
			// In production, you might want to implement retry logic or store failed events
		}
	}()

	// Return success immediately (non-blocking response)
	return c.JSON(fiber.Map{
		"status":    "success",
		"message":   "click logged",
		"ad_id":     input.AdID,
		"timestamp": event.Timestamp,
	})
}

// getClientIP extracts the real client IP from various headers
func getClientIP(c *fiber.Ctx) string {
	// Check X-Forwarded-For header (most common for proxies)
	if xff := c.Get("X-Forwarded-For"); xff != "" {
		// X-Forwarded-For can contain multiple IPs, take the first one
		if ips := strings.Split(xff, ","); len(ips) > 0 {
			ip := strings.TrimSpace(ips[0])
			if net.ParseIP(ip) != nil {
				return ip
			}
		}
	}

	// Check X-Real-IP header
	if xri := c.Get("X-Real-IP"); xri != "" {
		if net.ParseIP(xri) != nil {
			return xri
		}
	}

	// Fallback to Fiber's IP method
	return c.IP()
}