package handlers

import (
	"encoding/json"
	"log"
	"producer/kafka"
	"producer/utils"
	"time"

	// "video-ads-backend/producer/kafka"
	// "video-ads-backend/producer/utils"

	"github.com/gofiber/fiber/v2"
)

type ClickEvent struct {
    AdID            string    `json:"ad_id"`
    Timestamp       time.Time `json:"timestamp"`
    IP              string    `json:"ip"`
    PlaybackSeconds int       `json:"playback_seconds"`
}

func HandleAdClick(c *fiber.Ctx) error {
    var input struct {
        AdID            string `json:"ad_id"`
        PlaybackSeconds int    `json:"playback_seconds"`
    }

    if err := c.BodyParser(&input); err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request"})
    }

    event := ClickEvent{
        AdID:            input.AdID,
        Timestamp:       time.Now().UTC(),
        IP:              c.IP(),
        PlaybackSeconds: input.PlaybackSeconds,
    }

    data, err := json.Marshal(event)
    if err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to serialize event"})
    }

    cfg := utils.LoadConfig()
    if err := kafka.PublishMessage(cfg.KafkaBroker, cfg.KafkaTopic, data); err != nil {
        log.Printf("Failed to publish message: %v", err)
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to log click"})
    }

    return c.JSON(fiber.Map{"status": "click logged"})
}
