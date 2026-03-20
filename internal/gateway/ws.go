package gateway

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/alvisLu/go-shorten/internal/config"
	"github.com/alvisLu/go-shorten/internal/ws"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

const (
	readLimit    = 512
	pongWait     = 60 * time.Second
	pingInterval = 45 * time.Second
	writeWait    = 10 * time.Second
)

func newUpgrader(cfg *config.Config) websocket.Upgrader {
	allowed := make(map[string]struct{}, len(cfg.AllowedOrigins))
	for _, o := range cfg.AllowedOrigins {
		allowed[o] = struct{}{}
	}
	return websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			origin := r.Header.Get("Origin")
			if origin == "" {
				return true // non-browser clients (Postman, wscat, etc.)
			}
			_, ok := allowed[origin]
			return ok
		},
	}
}

type wsDispatcher map[string]func() (any, error)

func newWsDispatcher(svc ws.WsService) wsDispatcher {
	return wsDispatcher{
		"health": func() (any, error) { return svc.WsHealth(), nil },
	}
}

// readPump reads messages from the client and dispatches them to the send channel.
// It also handles pong messages to reset the read deadline.
func readPump(conn *websocket.Conn, dispatch wsDispatcher, send chan<- any, done chan<- struct{}) {
	defer close(done)

	conn.SetReadLimit(readLimit)
	conn.SetReadDeadline(time.Now().Add(pongWait))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, data, err := conn.ReadMessage()
		if err != nil {
			break
		}
		conn.SetReadDeadline(time.Now().Add(pongWait))

		var msg ws.WsReq
		if err := json.Unmarshal(data, &msg); err != nil {
			break
		}

		if handler, ok := dispatch[msg.Type]; ok {
			resp, err := handler()
			if err != nil {
				break
			}
			send <- resp
		}
	}
}

// writePump writes messages from the send channel to the client.
// It sends periodic pings to keep the connection alive.
func writePump(conn *websocket.Conn, send <-chan any, done <-chan struct{}) {
	ticker := time.NewTicker(pingInterval)
	defer ticker.Stop()

	for {
		select {
		case msg := <-send:
			conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := conn.WriteJSON(msg); err != nil {
				return
			}
		case <-ticker.C: // send ping
			conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		case <-done:
			return
		}
	}
}

func wsHandler(upgrader websocket.Upgrader) gin.HandlerFunc {
	return func(c *gin.Context) {
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		defer conn.Close()

		send := make(chan any, 8)
		done := make(chan struct{})

		go readPump(conn, newWsDispatcher(ws.NewService()), send, done)
		writePump(conn, send, done)
	}
}

func registerWsRoutes(cfg *config.Config, r *gin.Engine) {
	r.GET("/ws", wsHandler(newUpgrader(cfg)))
}
