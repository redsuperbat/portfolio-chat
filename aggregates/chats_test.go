package aggregates

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"rsb.asuscomm.com/portfolio-chat/events"
)

func TestChatsSet(t *testing.T) {
	chats := Chats{}
	chat := &Chat{
		Id:       "123",
		Messages: []*Message{},
	}
	chats.Set(chat)
	assert.NotNil(t, chats.Get("123"))
}

func TestChatsOn(t *testing.T) {
	chats := Chats{}
	chatStartedEvent := events.ChatStartedEvent{
		EventType: "ChatStartedEvent",
		ChatId:    "123",
	}
	chats.On(&chatStartedEvent)
	assert.NotNil(t, chats.Get("123"))
	chatMessageSentEvent := events.ChatMessageSentEvent{
		EventType: "ChatMessageSentEvent",
		ChatId:    "123",
		Content:   "Hello World",
		SenderId:  "123",
		SentAt:    time.Now().UTC(),
		MessageId: "123",
	}
	chats.On(&chatMessageSentEvent)
	assert.Equal(t, chats.Get("123").Messages[0].Id, "123")
}
