package events

import (
	"encoding/json"
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

func (c *ChatStartedEvent) Type() string {
	return c.EventType
}

func (c *ChatStartedEvent) ToBytes() ([]byte, error) {
	return json.Marshal(c)
}
