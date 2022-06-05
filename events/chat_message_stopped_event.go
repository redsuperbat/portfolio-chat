package events

import (
	"encoding/json"
	"errors"

	"github.com/google/uuid"
)

type ChatMessageStoppedEvent struct {
	EventType string `json:"eventType"`
	ChatId    string `json:"chatId"`
	SenderId  string `json:"senderId"`
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

func (c *ChatMessageStoppedEvent) Type() string {
	return c.EventType
}

func (c *ChatMessageStoppedEvent) ToBytes() ([]byte, error) {
	return json.Marshal(c)
}
