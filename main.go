package main

import (
	"log"

	gptAPI "github.com/EddyTeddy00/gpt-bot/gpt_3_5_turbo"
	tgAPI "github.com/go-telegram-bot-api/telegram-bot-api"
)

func main() {
	// API_TOKENS
	tokenGPT := ""
	tokenTG := "5848185673:AAGUArDsWxeMW-su4YsiP1Cgo2sT9VJ52No"

	// Telegram initialization
	bot, err := tgAPI.NewBotAPI(tokenTG)
	if err != nil {
		log.Panic(err)
	}

	// Enable bot API debug mode
	bot.Debug = true

	// Print out bot account information
	log.Printf("Authorized on account %s", bot.Self.UserName)

	// Chat GPT initialization
	gptbot, err := gptAPI.Init(gptAPI.Params{
		API_TOKEN:          tokenGPT,
		KeepMessageHistory: true,
		StripNewline:       true,
		Request: gptAPI.ChatRequest{
			Model: "gpt-3.5-turbo-0301",
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	// Start Telegram long polling update
	updateConfig := tgAPI.NewUpdate(0)
	updateConfig.Timeout = 60
	updates, err := bot.GetUpdatesChan(updateConfig)

	//Check updates for incoming messages from TG
	for update := range updates {
		// Skip any non-messages updates
		if update.Message == nil {
			continue
		}

		// Send message to ChatGPT
		choices, err := gptbot.Query(update.Message.Text)
		if err != nil {
			log.Println(err)
			continue
		}

		// Send response messages to Telegram
		for _, choice := range choices {
			msg := tgAPI.NewMessage(update.Message.Chat.ID, choice.Message.Content)
			msg.ReplyToMessageID = update.Message.MessageID

			_, err = bot.Send(msg)
			if err != nil {
				log.Println("Error:", err)
			}

		}
	}
}
