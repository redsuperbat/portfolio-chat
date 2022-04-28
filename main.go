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

type MsgChan = chan []byte

func handleEvent[T events.Event](c *fiber.Ctx, sendEvent func(value []byte) error) error {
	var event T
	if err := c.BodyParser(&event); err != nil {
		return c.Status(400).JSON(&fiber.Map{
			"message": "Invalid command",
			"code":    400,
		})
	}

	if err := event.Valid(); err != nil {
		return c.Status(400).JSON(&fiber.Map{
			"message": err.Error(),
			"code":    400,
		})
	}

	if err := sendEvent(c.Body()); err != nil {
		return c.Status(500).JSON(&fiber.Map{
			"message": err.Error(),
			"code":    500,
		})
	}

	return c.Status(200).JSON(event)
}

var allowedEvents map[string]bool = map[string]bool{
	"ChatMessageSentEvent":    true,
	"ChatMessageStartedEvent": true,
	"ChatMessageStoppedEvent": true,
}

func main() {

	sendEvent, closeProducer := producing.NewConn()
	defer closeProducer()
	app := fiber.New()
	app.Use(logger.New())
	app.Use(cors.New())

	ephemeralChannel := make(MsgChan)
	startEphCons, stopEphCons := consuming.NewConsumer(ephemeralChannel, uuid.NewString()+time.Now().String())
	log.Println("Starting ephemeral consumer...")
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
			chats.On(event)
		}
	}()

	app.Get("/chats/:id", func(c *fiber.Ctx) error {
		id := c.Params("id")
		chat := chats.Get(id)
		if chat == nil {
			return c.Status(404).JSON(fiber.Map{
				"message": fmt.Sprint("Chat with id ", id, " not found"),
				"code":    404,
			})
		}
		return c.Status(200).JSON(chat)
	})

	app.Post("/start-chat", func(c *fiber.Ctx) error {
		return handleEvent[*events.ChatStartedEvent](c, sendEvent)
	})

	app.Post("/send-chat-message", func(c *fiber.Ctx) error {
		return handleEvent[*events.ChatMessageSentEvent](c, sendEvent)
	})

	app.Post("/start-typing", func(c *fiber.Ctx) error {
		return handleEvent[*events.ChatMessageStartedEvent](c, sendEvent)
	})

	app.Post("/stop-typing", func(c *fiber.Ctx) error {
		return handleEvent[*events.ChatMessageStoppedEvent](c, sendEvent)
	})

	// ##### --------------- #####
	// ##### Websocket Stuff #####
	// ##### --------------- #####

	staticChan := make(MsgChan)
	wsChans := map[string]MsgChan{}
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
		log.Println("Id param:", chatId)
		senderId := c.Query("senderId")
		log.Println("Sender Id query:", senderId)

		wsChan := make(chan []byte)
		wsChans[senderId] = wsChan

		go func() {
			for msg := range wsChan {
				var msgMap fiber.Map
				if err := json.Unmarshal(msg, &msgMap); err != nil {
					log.Println(err.Error())
					break
				}
				if msgMap["chatId"] != c.Params("id") {
					continue
				}
				eventType := msgMap["eventType"].(string)
				if !allowedEvents[eventType] {
					continue
				}
				log.Printf("recv: %s", msgMap["eventType"])
				if err := c.WriteJSON(msgMap); err != nil {
					log.Println("write:", err)
					break
				}
			}
		}()

		for {
			if _, _, err := c.ReadMessage(); err != nil {
				log.Println("read:", err)
				fmt.Printf("Client %s disconnected from chat %s", senderId, chatId)
				close(wsChan)
				delete(wsChans, senderId)
				break
			}
		}

	}))

	log.Fatalln(app.Listen(":8080"))
}
