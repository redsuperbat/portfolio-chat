package events

import (
	"errors"

	"github.com/google/uuid"
)

type ChatMessageStoppedEvent struct {
	EventType string `json:"eventType"`
	ChatId    string `json:"chatId"`
	Sender    string `json:"sender"`
}

func (c *ChatMessageStoppedEvent) Valid() error {
	if c.EventType != "ChatMessageStoppedEvent" {
		return errors.New("Invalid event type")
	}
	if _, err := uuid.Parse(c.ChatId); err != nil {
		return errors.New("Invalid chat id")
	}
	return nil
}
