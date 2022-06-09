package aggregates

import (
	"time"

	"rsb.asuscomm.com/portfolio-chat/events"
)

type Message struct {
	Content  string    `json:"content"`
	SenderId string    `json:"senderId"`
	Id       string    `json:"messageId"`
	SentAt   time.Time `json:"sentAt"`
}

type ChatMember struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

type Chat struct {
	Id       string        `json:"chatId"`
	Members  []*ChatMember `json:"members"`
	Messages []*Message    `json:"messages"`
}

// Chats aggregate.
type Chats map[string]*Chat

func (c *Chats) On(event events.Event) {
	switch e := event.(type) {
	case *events.ChatStartedEvent:
		messages := []*Message{}
		chat := &Chat{
			Id:       e.ChatId,
			Messages: messages,
		}
		c.Set(chat)

	case *events.ChatMessageSentEvent:
		chat := c.Get(e.ChatId)
		if chat == nil {
			return
		}

		msg := &Message{
			Content:  e.Content,
			SenderId: e.SenderId,
			Id:       e.MessageId,
			SentAt:   e.SentAt,
		}
		chat.Messages = append(chat.Messages, msg)

	case *events.ChatJoinedEvent:
		chat := c.Get(e.ChatId)
		if chat == nil {
			return
		}
		chat.Members = append(chat.Members, &ChatMember{
			Id:   e.SenderId,
			Name: e.Name,
		})
	}
}

func (c Chats) Get(chatId string) *Chat {
	if val, ok := c[chatId]; ok {
		return val
	}
	return nil
}

func (c Chats) Set(chat *Chat) {
	c[chat.Id] = chat
}

func (c Chats) Has(chatId string) bool {
	return c[chatId] != nil
}

func (c Chats) HasName(name string) bool {
	for _, chat := range c {
		for _, member := range chat.Members {
			if member.Name == name {
				return true
			}
		}
	}
	return false
}
