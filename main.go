package main

import (
	"context"
	"log"

	"finance-bot/bot"
	"finance-bot/config"
	"finance-bot/db"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func main() {
	log.Println("Starting finance bot...")

	cfg := config.LoadConfig()
	log.Println("Config loaded successfully")

	dbClient := db.Connect(cfg.MongoURI, cfg.DBName)
	defer dbClient.Client.Disconnect(nil)
	log.Println("Connected to database")

	initializeDefaultCategory(dbClient)

	botHandler := bot.NewBotHandler(cfg.BotToken, dbClient, cfg)
	log.Println("Bot handler created")

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := botHandler.Bot.GetUpdatesChan(u)
	log.Println("Started receiving updates")

	for update := range updates {
		log.Println("Received new update")
		botHandler.HandleUpdate(update)
	}
}

func initializeDefaultCategory(dbClient *db.MongoDB) {
    filter := bson.M{"name": "Общая"}
    update := bson.M{
        "$setOnInsert": bson.M{
            "_id":     primitive.NewObjectID(),
            "user_id": 0,
            "name":    "Общая",
            "limit":   100000000000000.0,
        },
    }
    opts := options.Update().SetUpsert(true)
    _, err := dbClient.DB.Collection("categories").UpdateOne(context.Background(), filter, update, opts)
    if err != nil {
        log.Fatalf("Failed to initialize default category: %v", err)
    }
    log.Println("Default category initialized")
}
