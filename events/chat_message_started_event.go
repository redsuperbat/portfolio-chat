package events

import (
	"errors"

	"github.com/google/uuid"
)

type ChatMessageStartedEvent struct {
	EventType string `json:"eventType"`
	ChatId    string `json:"chatId"`
	Sender    string `json:"sender"`
}

func (c *ChatMessageStartedEvent) Valid() error {
	if c.EventType != "ChatMessageStartedEvent" {
		return errors.New("Invalid event type")
	}
	if _, err := uuid.Parse(c.ChatId); err != nil {
		return errors.New("Invalid chat id")
	}
	return nil
}
