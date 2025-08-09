package notifier

import (
	"fmt"
	tgbotapi "github.com/Syfaro/telegram-bot-api"
	"log"
)

type Notifier struct {
	bot *tgbotapi.BotAPI
}

func New(token string) (*Notifier, error) {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, fmt.Errorf("failed to create Telegram bot: %w", err)
	}

	log.Printf("Authorized on account %s", bot.Self.UserName)

	return &Notifier{bot: bot}, nil
}

func (n *Notifier) SendNotification(recipientID int64, text string) error {
	msg := tgbotapi.NewMessage(recipientID, text)
	_, err := n.bot.Send(msg)
	if err != nil {
		return fmt.Errorf("failed to send message to Telegram: %w", err)
	}

	return nil
}
