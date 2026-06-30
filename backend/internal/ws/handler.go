package ws

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"

	"github.com/MaximovIlya/vk_practice_project/internal/auth"
)

// Upgrade — middleware, проверяет что запрос является WebSocket upgrade
func Upgrade(c *fiber.Ctx) error {
	if websocket.IsWebSocketUpgrade(c) {
		return c.Next()
	}
	return fiber.ErrUpgradeRequired
}

// Handler — возвращает Fiber-хендлер для WebSocket соединений.
// Токен передаётся query-параметром: /ws?token=<access_token>
func Handler(hub *Hub, manager *auth.Manager) fiber.Handler {
	return websocket.New(func(conn *websocket.Conn) {
		token := conn.Query("token")
		claims, err := manager.Parse(token)
		if err != nil || claims.Type != auth.AccessToken {
			_ = conn.WriteMessage(websocket.CloseMessage,
				websocket.FormatCloseMessage(websocket.ClosePolicyViolation, "unauthorized"))
			return
		}

		client := newClient(conn, hub, claims.UserID)
		hub.register <- client

		go client.writePump()
		client.readPump() // блокирует до закрытия соединения
	})
}
