package ws

import (
	"encoding/json"
	"log"

	"github.com/gofiber/websocket/v2"
)

// Client — одно WebSocket-соединение
type Client struct {
	conn      *websocket.Conn
	hub       *Hub
	send      chan []byte
	userID    string
	sessionID string // заполняется при join-room / organizer-join
}

func newClient(conn *websocket.Conn, hub *Hub, userID string) *Client {
	return &Client{
		conn:   conn,
		hub:    hub,
		send:   make(chan []byte, 64),
		userID: userID,
	}
}

// readPump читает сообщения от клиента и передаёт в хаб
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
	}()

	for {
		_, raw, err := c.conn.ReadMessage()
		if err != nil {
			break
		}

		var msg InMessage
		if err := json.Unmarshal(raw, &msg); err != nil {
			log.Printf("[ws] bad message from %s: %v", c.userID, err)
			continue
		}

		c.hub.incoming <- clientMessage{client: c, msg: msg}
	}
}

// writePump отправляет сообщения из канала send клиенту
func (c *Client) writePump() {
	defer c.conn.Close()

	for data := range c.send {
		if err := c.conn.WriteMessage(websocket.TextMessage, data); err != nil {
			break
		}
	}
}

func (c *Client) emit(msgType string, payload any) {
	select {
	case c.send <- encode(msgType, payload):
	default:
		log.Printf("[ws] send buffer full for user %s", c.userID)
	}
}
