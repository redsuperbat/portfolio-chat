package aggregates

import (
	"time"

	"rsb.asuscomm.com/portfolio-chat/events"
)

type Message struct {
	Content string    `json:"content"`
	Sender  string    `json:"sender"`
	Id      string    `json:"messageId"`
	SentAt  time.Time `json:"sentAt"`
}

type Chat struct {
	Id       string     `json:"chatId"`
	Messages []*Message `json:"messages"`
}

// Patient aggregate.
type Chats map[string]*Chat

func (c *Chats) On(event events.Event) {
	switch e := event.(type) {
	case *events.ChatMessageSentEvent:
		chat := c.Get(e.ChatId)
		if chat == nil {
			return
		}

		msg := &Message{
			Content: e.Content,
			Sender:  e.Sender,
			Id:      e.MessageId,
			SentAt:  e.SentAt,
		}
		chat.Messages = append(chat.Messages, msg)

	case *events.ChatStartedEvent:
		messages := []*Message{}
		chat := &Chat{
			Id:       e.ChatId,
			Messages: messages,
		}
		c.Set(chat)
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

func NewFromEvents(events []events.Event) *Chats {
	p := &Chats{}

	for _, event := range events {
		p.On(event)
	}

	return p
}
