package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/websocket/v2"
	"github.com/google/uuid"

	"rsb.asuscomm.com/portfolio-chat/aggregates"
	"rsb.asuscomm.com/portfolio-chat/consuming"
	"rsb.asuscomm.com/portfolio-chat/events"
	"rsb.asuscomm.com/portfolio-chat/producing"
)

type ByteChannel = chan []byte

func errResp(code int, msg string) *fiber.Map {
	return &fiber.Map{
		"message": msg,
		"code":    code,
	}
}

func dispatchCommand(c *fiber.Ctx, sendEvent func(value []byte) error) error {
	if err := sendEvent(c.Body()); err != nil {
		return c.Status(500).JSON(errResp(500, err.Error()))
	}
	return c.SendStatus(204)
}

func validateCommand[T events.Event](c *fiber.Ctx) (error, T) {
	var event T
	if err := c.BodyParser(&event); err != nil {
		return c.Status(400).JSON(errResp(400, "Invalid Command")), event
	}

	if err := event.Valid(); err != nil {
		return c.Status(400).JSON(errResp(400, err.Error())), event
	}
	return nil, event
}

var allowedEvents map[string]bool = map[string]bool{
	"ChatMessageSentEvent":    true,
	"ChatMessageStartedEvent": true,
	"ChatMessageStoppedEvent": true,
	"ChatJoinedEvent":         true,
}

func main() {

	sendEvent, closeProducer := producing.NewConn()
	defer closeProducer()
	app := fiber.New()
	app.Use(logger.New())
	app.Use(cors.New())

	ephemeralChannel := make(ByteChannel)
	startEphCons, stopEphCons := consuming.NewConsumer(ephemeralChannel, uuid.NewString()+time.Now().String())
	go startEphCons()
	defer stopEphCons()

	chats := aggregates.Chats{}

	go func() {
		for val := range ephemeralChannel {
			event, err := events.Unmarshal(val)
			if err != nil {
				log.Printf("Unable to Unmarshal event %v because of %s", val, err.Error())
				continue
			}
			log.Printf("Handling event %s", event.Type())
			chats.On(event)
		}
	}()

	app.Get("/chats/:id", func(c *fiber.Ctx) error {
		id := c.Params("id")
		chat := chats.Get(id)

		if chat == nil {
			errMsg := fmt.Sprint("Chat with id ", id, " not found")
			return c.Status(404).JSON(errResp(404, errMsg))
		}

		return c.Status(200).JSON(chat)
	})

	app.Get("/chats/:id/members", func(c *fiber.Ctx) error {
		id := c.Params("id")
		chat := chats.Get(id)

		if chat == nil {
			errMsg := fmt.Sprint("Chat with id ", id, " not found")
			return c.Status(404).JSON(errResp(404, errMsg))
		}

		return c.Status(200).JSON(chat.Members)
	})

	app.Post("/start-chat", func(c *fiber.Ctx) error {
		var StartChatDto struct {
			ChosenName string `json:"chosenName"`
		}
		c.BodyParser(&StartChatDto)
		if chats.HasName(StartChatDto.ChosenName) {
			return c.Status(400).JSON(errResp(400, "Name taken"))
		}
		chatId := uuid.NewString()
		senderId := uuid.NewString()
		chatStartedEvent := &events.ChatStartedEvent{
			ChatId:    chatId,
			EventType: "ChatStartedEvent",
		}
		chatJoinedEvent := &events.ChatJoinedEvent{
			ChatId:    chatId,
			SenderId:  senderId,
			EventType: "ChatJoinedEvent",
			Name:      StartChatDto.ChosenName,
		}
		for _, val := range []events.Event{chatStartedEvent, chatJoinedEvent} {
			bytes, err := val.ToBytes()
			if err != nil {
				return c.SendStatus(500)
			}
			sendEvent(bytes)
		}
		return c.Status(201).JSON(&fiber.Map{
			"chatId":   chatId,
			"senderId": senderId,
		})
	})

	app.Post("/join-chat", func(c *fiber.Ctx) error {
		var JoinChatDto struct {
			ChatId string `json:"chatId"`
			Name   string `json:"name"`
		}
		if err := c.BodyParser(&JoinChatDto); err != nil {
			return c.SendStatus(500)
		}
		SenderId := uuid.NewString()
		chatJoinedEvent := &events.ChatJoinedEvent{
			EventType: "ChatJoinedEvent",
			ChatId:    JoinChatDto.ChatId,
			SenderId:  SenderId,
			JoinedAt:  time.Now(),
			Name:      JoinChatDto.Name,
		}
		bytes, err := chatJoinedEvent.ToBytes()
		if err != nil {
			return c.SendStatus(500)
		}
		err = sendEvent(bytes)
		if err != nil {
			return c.SendStatus(500)
		}
		return c.Status(201).JSON(&fiber.Map{
			"senderId": SenderId,
		})
	})

	app.Post("/send-chat-message", func(c *fiber.Ctx) error {
		var SendChatMessageDto struct {
			ChatId   string    `json:"chatId"`
			Content  string    `json:"content"`
			SenderId string    `json:"senderId"`
			SentAt   time.Time `json:"sentAt"`
		}
		c.BodyParser(&SendChatMessageDto)

		chatMessageSentEvent := &events.ChatMessageSentEvent{
			EventType: "ChatMessageSentEvent",
			ChatId:    SendChatMessageDto.ChatId,
			Content:   SendChatMessageDto.Content,
			SentAt:    SendChatMessageDto.SentAt,
			MessageId: uuid.NewString(),
			SenderId:  SendChatMessageDto.SenderId,
		}
		bytes, err := chatMessageSentEvent.ToBytes()
		if err != nil {
			return c.SendStatus(500)
		}
		err = sendEvent(bytes)
		if err != nil {
			return c.SendStatus(500)
		}
		return c.SendStatus(201)
	})

	app.Post("/start-typing", func(c *fiber.Ctx) error {
		if err, _ := validateCommand[*events.ChatMessageStartedEvent](c); err != nil {
			return err
		}
		return dispatchCommand(c, sendEvent)
	})

	app.Post("/stop-typing", func(c *fiber.Ctx) error {
		if err, _ := validateCommand[*events.ChatMessageStoppedEvent](c); err != nil {
			return err
		}
		return dispatchCommand(c, sendEvent)
	})

	// ##### --------------- #####
	// ##### Websocket Stuff #####
	// ##### --------------- #####

	staticChan := make(ByteChannel)
	wsChans := make(map[string]ByteChannel)
	startStaticCons, stopStaticCons := consuming.NewConsumer(staticChan, "ChatMessageConsumer")
	go startStaticCons()
	defer stopStaticCons()

	go func() {
		for val := range staticChan {
			for _, v := range wsChans {
				v <- val
			}
		}
	}()

	app.Use("/chat-room", func(c *fiber.Ctx) error {

		// IsWebSocketUpgrade returns true if the client
		// requested upgrade to the WebSocket protocol.
		if websocket.IsWebSocketUpgrade(c) {
			c.Locals("allowed", true)
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})

	app.Get("/chat-room/:id", websocket.New(func(c *websocket.Conn) {
		log.Println("Allowed:", c.Locals("allowed"))
		chatId := c.Params("id")

		if _, err := uuid.Parse(chatId); err != nil {
			log.Printf("Invalid chat id %s", chatId)
			c.Close()
			return
		}
		senderId := c.Query("senderId")

		log.Println("Id param:", chatId)

		if _, err := uuid.Parse(senderId); err != nil {
			log.Printf("Invalid sender id %s", senderId)
			c.Close()
			return
		}

		log.Println("Sender Id query:", senderId)
		wsChan := make(ByteChannel)
		wsChans[senderId] = wsChan

		log.Printf("Client connected! Number of clients: %v", len(wsChans))

		go func() {
			for msg := range wsChan {
				var msgMap fiber.Map
				if err := json.Unmarshal(msg, &msgMap); err != nil {
					log.Println(err.Error())
					break
				}
				// Checks if the message is in the current chat scope.
				if msgMap["chatId"] != c.Params("id") {
					continue
				}
				eventType := msgMap["eventType"].(string)
				if !allowedEvents[eventType] {
					continue
				}
				log.Printf("Sending %s to client %s", msgMap["eventType"], senderId)
				if err := c.WriteJSON(msgMap); err != nil {
					log.Println("write:", err)
					break
				}
			}
		}()

		for {
			if _, _, err := c.ReadMessage(); err != nil {
				log.Println("read:", err)
				log.Println("Closing channel")
				close(wsChan)
				log.Println("Deallocating resources for that channel")
				delete(wsChans, senderId)
				log.Printf("Client %s disconnected from chat %s. Number of clients: %v\n", senderId, chatId, len(wsChans))
				break
			}
		}

	}))

	log.Fatalln(app.Listen(":8080"))
}
