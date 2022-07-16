package application

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"
)

func HandleSocketEvents(ctx context.Context, client *slack.Client, socketClient *socketmode.Client) {
	for {
		select {
		case <-ctx.Done():
			log.Println("Shutting down socketmode listener")
			return

		case event := <-socketClient.Events:
			switch event.Type {
			case socketmode.EventTypeEventsAPI:
				eventsAPIEvent, ok := event.Data.(slackevents.EventsAPIEvent)
				if !ok {
					log.Printf("Could not type cast the event to the EventsAPIEvent: %v\n", event)
					continue
				}

				socketClient.Ack(*event.Request)
				err := handleEventMessage(eventsAPIEvent, client)
				if err != nil {
					log.Fatal(err)
				}
			case socketmode.EventTypeSlashCommand:
				command, ok := event.Data.(slack.SlashCommand)
				if !ok {
					log.Printf("Could not type cast the message to a SlashCommand: %v\n", command)
					continue
				}
				//acknowledge the event
				socketClient.Ack(*event.Request)
				err := handleSlashCommand(command, client)
				if err != nil {
					log.Fatal(err)
				}
			}
		}
	}
}

func handleEventMessage(event slackevents.EventsAPIEvent, client *slack.Client) error {
	switch event.Type {
	// First we check if this is an CallbackEvent
	case slackevents.CallbackEvent:

		innerEvent := event.InnerEvent
		// Yet Another Type switch on the actual Data to see if its an AppMentionEvent
		switch ev := innerEvent.Data.(type) {
		case *slackevents.AppMentionEvent:
			// The application has been mentioned since this Event is a Mention event
			user, err := client.GetUserInfo(ev.User)
			if err != nil {
				return err
			}
			if user.IsBot {
				return nil
			}
			err = handleAppMentionEvent(ev, user, client)
			if err != nil {
				return err
			}
		case *slackevents.MessageEvent:
			user, err := client.GetUserInfo(ev.User)
			if err != nil {
				return err
			}
			if user.IsBot {
				return nil
			}
			err = handleMessageEvent(ev, user, client)
			if err != nil {
				return err
			}
		}
	default:
		return errors.New("unsupported event type")
	}
	return nil
}

func handleAppMentionEvent(event *slackevents.AppMentionEvent, user *slack.User, client *slack.Client) error {

	// Grab the user name based on the ID of the one who mentioned the bot
	user, err := client.GetUserInfo(event.User)
	if err != nil {
		return err
	}
	// Check if the user said Hello to the bot
	text := strings.ToLower(event.Text)

	// Create the attachment and assigned based on the message
	attachment := slack.Attachment{}
	// Add Some default context like user who mentioned the bot
	attachment.Fields = []slack.AttachmentField{
		{
			Title: "Date",
			Value: time.Now().String(),
		}, {
			Title: "Initializer",
			Value: user.Name,
		},
	}
	if strings.Contains(text, "hello") {
		// Greet the user
		attachment.Text = fmt.Sprintf("Hello %s", user.Name)
		attachment.Pretext = "Greetings"
		attachment.Color = "#4af030"
	} else {
		// Send a message to the user
		attachment.Text = fmt.Sprintf("How can I help you %s?", user.Name)
		attachment.Pretext = "How can I be of service"
		attachment.Color = "#3d3d3d"
	}
	// Send the message to the channel
	// The Channel is available in the event message
	_, _, err = client.PostMessage(event.Channel, slack.MsgOptionAttachments(attachment))
	if err != nil {
		return fmt.Errorf("failed to post message: %w", err)
	}
	return nil
}

func handleMessageEvent(event *slackevents.MessageEvent, user *slack.User, client *slack.Client) error {

	// Grab the user name based on the ID of the one who mentioned the bot
	user, err := client.GetUserInfo(event.User)
	if err != nil {
		return err
	}
	if user.IsBot {
		return nil
	}
	// Check if the user said Hello to the bot
	text := strings.ToLower(event.Text)

	// Create the attachment and assigned based on the message
	var message string
	// Add Some default context like user who mentioned the bot
	if strings.Contains(text, "hello") {
		// Greet the user
		message = fmt.Sprintf("Hello %s, how can I help you get unstuck?", user.Name)
	} else if strings.Contains(text, "thank you"){
		message = fmt.Sprintf("You are welcome %s", user.Name)
	} else {
		// Send a message to the user
		var helpMessages = []string{"Did you check the logs?", "What did you do to debug?", "Do not worry, you won't be stuck on this forever.", "Maybe take a break from this one and look at it again later."}
		rand.Seed(time.Now().UnixNano())
		message = helpMessages[rand.Intn(len(helpMessages))]
	}
	// Send the message to the channel
	// The Channel is available in the event message
	_, _, err = client.PostMessage(event.Channel, slack.MsgOptionText(message, false))
	if err != nil {
		return fmt.Errorf("failed to post message: %w", err)
	}
	return nil
}

func handleSlashCommand(command slack.SlashCommand, client *slack.Client) error {
	switch command.Command {
	case "/lookup":
		return handleLookUpCommand(command, client)
	}
	return nil
}

func handleLookUpCommand(command slack.SlashCommand, client *slack.Client) error {
	message := command.Text
	splitStr := strings.Split(message, " ")

	parsed := strings.Join(splitStr, `-`)
	url := fmt.Sprintf(`https://api.stackexchange.com/2.3/search/advanced?pagesize=3&order=desc&sort=relevance&q=%s&site=stackoverflow`, parsed)
	res, err := http.Get(url)
	if err != nil {
		log.Fatalln(err)
		return err
	}

	var resp StackResponse
	json.NewDecoder(res.Body).Decode(&resp)

	for _, v := range resp.Items {
		m := fmt.Sprintf(`%s
%s`, v.Title, v.Link)
		_, _, err = client.PostMessage(command.ChannelID, slack.MsgOptionText(m, false))
		if err != nil {
			log.Fatalln("failed to post message: %w", err)
			return err
		}
		time.Sleep(5 * time.Second)
	}
	return nil
}
