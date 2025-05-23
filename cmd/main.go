package main

import (
	"github.com/digkill/consultantCyberCTO/internal/handlers"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"

	"log"
	"os"
)

func main() {

	if err := godotenv.Load(); err != nil {
		log.Panic("load env failed: %v", err)
	}

	bot, err := tgbotapi.NewBotAPI(os.Getenv("TELEGRAM_BOT_TOKEN"))
	if err != nil {
		log.Panic(err)
	}

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil {
			go handlers.HandleMessage(bot, update.Message)
		}
	}
}
