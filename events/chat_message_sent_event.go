package events

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
)

type ChatMessageSentEvent struct {
	EventType string    `json:"eventType"`
	ChatId    string    `json:"chatId"`
	Content   string    `json:"content"`
	SenderId  string    `json:"senderId"`
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

func (c *ChatMessageSentEvent) Type() string {
	return c.EventType
}

func (c *ChatMessageSentEvent) ToBytes() ([]byte, error) {
	return json.Marshal(c)
}
