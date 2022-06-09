package events

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
)

type ChatJoinedEvent struct {
	EventType string    `json:"eventType"`
	ChatId    string    `json:"chatId"`
	SenderId  string    `json:"senderId"`
	JoinedAt  time.Time `json:"joinedAt"`
	Name      string    `json:"name"`
}

func (c *ChatJoinedEvent) Valid() error {
	if c.EventType != "ChatJoinedEvent" {
		return errors.New("Invalid event type")
	}
	if _, err := uuid.Parse(c.ChatId); err != nil {
		return errors.New("Invalid chat id")
	}
	return nil
}

func (c *ChatJoinedEvent) Type() string {
	return c.EventType
}

func (c *ChatJoinedEvent) ToBytes() ([]byte, error) {
	return json.Marshal(c)
}
