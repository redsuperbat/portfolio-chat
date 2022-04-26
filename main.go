package main

import (
	"encoding/json"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/websocket/v2"
	"rsb.asuscomm.com/portfolio-chat/consuming"
	"rsb.asuscomm.com/portfolio-chat/events"
	"rsb.asuscomm.com/portfolio-chat/producing"
)

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

func main() {

	sendEvent, close := producing.NewConn()
	defer close()
	app := fiber.New()
	app.Use(logger.New())
	app.Use(cors.New())

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
		// c.Locals is added to the *websocket.Conn
		log.Println("Allowed:", c.Locals("allowed")) // true
		log.Println("id param:", c.Params("id"))     // 123
		channel := make(chan []byte)
		start, stop := consuming.NewConsumer(channel)
		go start()

		c.SetCloseHandler(func(code int, text string) error {
			return stop()
		})

		allowedEvents := map[string]bool{
			"ChatMessageSentEvent":    true,
			"ChatMessageStartedEvent": true,
			"ChatMessageStoppedEvent": true,
		}

		for msg := range channel {
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

	}))

	log.Fatalln(app.Listen(":8080"))
}
