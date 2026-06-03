package websocket

import (
	"log"

	fiberws "github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
)

type Handler struct {
	hub *Hub
}

func NewHandler(hub *Hub) *Handler {
	return &Handler{hub: hub}
}

func (h *Handler) RegisterRoutes(router fiber.Router) {
	router.Use("/ws", func(c *fiber.Ctx) error {
		if fiberws.IsWebSocketUpgrade(c) {
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})

	router.Get("/ws", fiberws.New(func(conn *fiberws.Conn) {
		h.hub.Add(conn)
		defer h.hub.Remove(conn)

		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				log.Printf("websocket disconnected: %v", err)
				return
			}
		}
	}))
}
