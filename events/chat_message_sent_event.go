package events

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type ChatMessageSentEvent struct {
	EventType string    `json:"eventType"`
	ChatId    string    `json:"chatId"`
	Content   string    `json:"content"`
	Sender    string    `json:"sender"`
	SentAt    time.Time `json:"sentAt"`
	MessageId string    `json:"messageId"`
}

func (c *ChatMessageSentEvent) Valid() error {
	if c.EventType != "ChatMessageSentEvent" {
		return errors.New("Invalid event type")
	}
	if _, err := uuid.Parse(c.ChatId); err != nil {
		return errors.New("Invalid chat id")
	}
	return nil
}
