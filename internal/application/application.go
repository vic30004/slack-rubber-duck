package application

import (
	"context"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/socketmode"
)

func Start() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	appToken := os.Getenv("BOT_AUTH_TOKEN")
	token := os.Getenv("WEBSOKET_TOKEN")
	slackAPI := slack.New(appToken, slack.OptionDebug(true), slack.OptionAppLevelToken(token))
	socketClient := socketmode.New(
		slackAPI,
		socketmode.OptionDebug(true),
		// Option to set a custom logger
		socketmode.OptionLog(log.New(os.Stdout, "socketmode: ", log.Lshortfile|log.LstdFlags)),
	)
	// Create a context that can be used to cancel goroutine
	ctx, cancel := context.WithCancel(context.Background())
	// Make this cancel called properly in a real program , graceful shutdown etc
	defer cancel()

	go HandleSocketEvents(ctx, slackAPI, socketClient)

	socketClient.Run()
}
