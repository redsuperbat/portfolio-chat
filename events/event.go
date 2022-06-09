package events

import (
	"encoding/json"
	"errors"
)

type Event interface {
	Valid() error
	Type() string
	ToBytes() ([]byte, error)
}

func Unmarshal(bytes []byte) (Event, error) {
	var baseEvent struct {
		EventType string `json:"eventType"`
	}
	if err := json.Unmarshal(bytes, &baseEvent); err != nil {
		return nil, err
	}

	switch baseEvent.EventType {
	case "ChatMessageSentEvent":
		var event ChatMessageSentEvent
		return &event, json.Unmarshal(bytes, &event)

	case "ChatStartedEvent":
		var event ChatStartedEvent
		return &event, json.Unmarshal(bytes, &event)

	case "ChatMessageStartedEvent":
		var event ChatMessageStartedEvent
		return &event, json.Unmarshal(bytes, &event)

	case "ChatMessageStoppedEvent":
		var event ChatMessageStoppedEvent
		return &event, json.Unmarshal(bytes, &event)

	case "ChatJoinedEvent":
		var event ChatJoinedEvent
		return &event, json.Unmarshal(bytes, &event)
	}

	return nil, errors.New("Undefined event")
}
