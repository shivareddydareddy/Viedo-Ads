package services

import (
	"consumer/db"
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
)

type ClickEvent struct {
	AdID            string    `json:"ad_id"`
	Timestamp       time.Time `json:"timestamp"`
	IP              string    `json:"ip"`
	PlaybackSeconds int       `json:"playback_seconds"`
	UserAgent       string    `json:"user_agent,omitempty"`
}

type MongoClient struct {
	Client   *mongo.Client
	Database *mongo.Database
}


func ProcessClickEvent(event ClickEvent, mongoClient *db.MongoClient, redisClient *db.RedisClient) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	log.Printf("Processing click event for AdID: %s", event.AdID)

	if err := storeToMongoDB(ctx, event, mongoClient); err != nil {
		log.Printf("MongoDB storage failed: %v", err)

	}


	if err := updateRedisAnalytics(ctx, event, redisClient); err != nil {
		log.Printf("Redis analytics update failed: %v", err)
		return err
	}

	return nil
}

func storeToMongoDB(ctx context.Context, event ClickEvent, mongoClient *db.MongoClient) error {
	collection := mongoClient.Database.Collection("click_events")
	
	document := map[string]interface{}{
		"ad_id":             event.AdID,
		"timestamp":         event.Timestamp,
		"ip":                event.IP,
		"playback_seconds":  event.PlaybackSeconds,
		"user_agent":        event.UserAgent,
		"processed_at":      time.Now().UTC(),
	}

	result, err := collection.InsertOne(ctx, document)
	if err != nil {
		return fmt.Errorf("failed to insert into MongoDB: %w", err)
	}

	log.Printf("Stored click event in MongoDB with ID: %v", result.InsertedID)
	return nil
}

func updateRedisAnalytics(ctx context.Context, event ClickEvent, redisClient *db.RedisClient) error {
	pipe := redisClient.Client.Pipeline()

	totalClicksKey := fmt.Sprintf("clicks:total:%s", event.AdID)
	pipe.Incr(ctx, totalClicksKey)
	pipe.Expire(ctx, totalClicksKey, 24*time.Hour)
	minuteKey := fmt.Sprintf("clicks:minute:%s:%s", 
		event.AdID, 
		event.Timestamp.Format("200601021504")) // YYYYMMDDHHMM
	pipe.Incr(ctx, minuteKey)
	pipe.Expire(ctx, minuteKey, 1*time.Hour) 
	hourKey := fmt.Sprintf("clicks:hour:%s:%s", 
		event.AdID, 
		event.Timestamp.Format("2006010215")) // YYYYMMDDHH
	pipe.Incr(ctx, hourKey)
	pipe.Expire(ctx, hourKey, 24*time.Hour) // Keep for 24 hours

	// 4. Track daily clicks
	dayKey := fmt.Sprintf("clicks:day:%s:%s", 
		event.AdID, 
		event.Timestamp.Format("20060102")) // YYYYMMDD
	pipe.Incr(ctx, dayKey)
	pipe.Expire(ctx, dayKey, 7*24*time.Hour) // Keep for 7 days

	// 5. Add to recent clicks sorted set
	recentClicksKey := fmt.Sprintf("clicks:recent:%s", event.AdID)
	pipe.ZAdd(ctx, recentClicksKey, redis.Z{
		Score:  float64(event.Timestamp.Unix()),
		Member: fmt.Sprintf("%s-%d", event.IP, event.Timestamp.UnixNano()),
	})
	pipe.Expire(ctx, recentClicksKey, 1*time.Hour)
	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to update: %w", err)
	}

	log.Printf("Updated Redis analytics for AdID: %s", event.AdID)
	return nil
}



type AdAnalytics struct {
	AdID         string    `json:"ad_id"`
	TotalClicks  int       `json:"total_clicks"`
	RecentClicks int       `json:"recent_clicks"`
	CTR          float64   `json:"ctr_percentage"`
	Timestamp    time.Time `json:"timestamp"`
}

func GetAdAnalytics(mongoClient *db.MongoClient,adID string, timeWindow time.Duration, redisClient *db.RedisClient) (*AdAnalytics, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	analytics := &AdAnalytics{
		AdID:      adID,
		Timestamp: time.Now().UTC(),
	}

	// Get total clicks
	totalClicksKey := fmt.Sprintf("clicks:total:%s", adID)
	totalClicksStr, err := redisClient.Client.Get(ctx, totalClicksKey).Result()
	if err != nil && err != redis.Nil {
		return nil, fmt.Errorf("failed to get total clicks: %w", err)
	}
	if totalClicksStr != "" {
		analytics.TotalClicks, _ = strconv.Atoi(totalClicksStr)
	}

	// Get recent clicks from sorted set
	recentClicksKey := fmt.Sprintf("clicks:recent:%s", adID)
	now := time.Now()
	fromTime := now.Add(-timeWindow)
	
	recentCount, err := redisClient.Client.ZCount(ctx, recentClicksKey, 
		strconv.FormatInt(fromTime.Unix(), 10),
		strconv.FormatInt(now.Unix(), 10)).Result()
	if err != nil && err != redis.Nil {
		return nil, fmt.Errorf("failed to get recent clicks: %w", err)
	}
	analytics.RecentClicks = int(recentCount)
	impressions := 1000
	if impressions > 0 {
		analytics.CTR = float64(analytics.TotalClicks) / float64(impressions) * 100
	}

	return analytics, nil
}
