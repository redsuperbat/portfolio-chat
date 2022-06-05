package aggregates

import (
	"time"

	"rsb.asuscomm.com/portfolio-chat/events"
)

type Message struct {
	Content  string    `json:"content"`
	SenderId string    `json:"sender"`
	Id       string    `json:"messageId"`
	SentAt   time.Time `json:"sentAt"`
}

type Chat struct {
	Id         string     `json:"chatId"`
	SenderName string     `json:"senderName"`
	SenderId   string     `json:"senderId"`
	Messages   []*Message `json:"messages"`
}

// Chats aggregate.
type Chats map[string]*Chat

func (c *Chats) On(event events.Event) {
	switch e := event.(type) {
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

	case *events.ChatStartedEvent:
		messages := []*Message{}
		chat := &Chat{
			Id:       e.ChatId,
			Messages: messages,
		}
		c.Set(chat)
	case *events.NameChosenEvent:
		event, ok := event.(*events.NameChosenEvent)
		if !ok {
			return
		}
		chat := c.Get(event.ChatId)
		if chat == nil {
			return
		}
		chat.SenderName = event.ChosenName
		chat.SenderId = event.SenderId
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
	for _, val := range c {
		if val.SenderName == name {
			return true
		}
	}
	return false
}
