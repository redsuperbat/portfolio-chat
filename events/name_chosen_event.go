package events

import (
	"encoding/json"
	"errors"

	"github.com/google/uuid"
)

type NameChosenEvent struct {
	EventType  string `json:"eventType"`
	ChatId     string `json:"chatId"`
	SenderId   string `json:"senderId"`
	ChosenName string `json:"chosenName"`
}

func (c *NameChosenEvent) Valid() error {
	if c.EventType != "ChatStartedEvent" {
		return errors.New("Invalid event type")
	}
	if _, err := uuid.Parse(c.ChatId); err != nil {
		return errors.New("Invalid chat id")
	}
	return nil
}

func (c *NameChosenEvent) Type() string {
	return c.EventType
}

func (c *NameChosenEvent) ToBytes() ([]byte, error) {
	return json.Marshal(c)
}
