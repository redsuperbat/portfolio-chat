package events

import (
	"errors"

	"github.com/google/uuid"
)

type ChatStartedEvent struct {
	EventType string `json:"eventType"`
	ChatId    string `json:"chatId"`
}

func (c *ChatStartedEvent) Valid() error {
	if c.EventType != "ChatStartedEvent" {
		return errors.New("Invalid event type")
	}
	if _, err := uuid.Parse(c.ChatId); err != nil {
		return errors.New("Invalid chat id")
	}
	return nil
}
