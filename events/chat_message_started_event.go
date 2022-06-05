package events

import (
	"encoding/json"
	"errors"

	"github.com/google/uuid"
)

type ChatMessageStartedEvent struct {
	EventType string `json:"eventType"`
	ChatId    string `json:"chatId"`
	SenderId  string `json:"senderId"`
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

func (c *ChatMessageStartedEvent) Type() string {
	return c.EventType
}

func (c *ChatMessageStartedEvent) ToBytes() ([]byte, error) {
	return json.Marshal(c)
}
