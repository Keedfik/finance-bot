package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	BotToken string
	MongoURI string
	DBName   string
}

func LoadConfig() *Config {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	return &Config{
		BotToken: os.Getenv("BOT_TOKEN"),
		MongoURI: os.Getenv("MONGO_URI"),
		DBName:   os.Getenv("DB_NAME"),
	}
}
