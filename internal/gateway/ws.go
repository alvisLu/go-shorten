package gateway

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/alvisLu/go-shorten/internal/config"
	"github.com/alvisLu/go-shorten/internal/session"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

const (
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

type controlMsg struct {
	Type       string `json:"type"`
	SourceLang string `json:"sourceLang"`
	TargetLang string `json:"targetLang"`
	SampleRate int    `json:"sampleRate"`
}

func handleControl(sess *session.Session, raw []byte) {
	var msg controlMsg
	if err := json.Unmarshal(raw, &msg); err != nil {
		return
	}
	switch msg.Type {
	case "start":
		sess.Start(msg.SourceLang, msg.TargetLang, msg.SampleRate)
	case "stop":
		sess.Stop()
	case "health":
		sess.Health()
	default:
		sess.Send(session.WsErrorResp{Error: "unknown control type"})
	}
}

func handleAudio(sess *session.Session, data []byte) {
	// Phase 5: parse binary frame and dispatch to ChannelState
	// [isFinal: 1 byte][channel: 1 byte][id: 21 bytes][PCM: N bytes Float32LE]
}

// readPump reads messages from the client and dispatches them.
// It also handles pong messages to reset the read deadline.
func readPump(conn *websocket.Conn, sess *session.Session, done chan<- struct{}, quit <-chan struct{}) {
	defer close(done)

	conn.SetReadDeadline(time.Now().Add(pongWait))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		msgType, data, err := conn.ReadMessage()
		if err != nil {
			break
		}
		conn.SetReadDeadline(time.Now().Add(pongWait))

		switch msgType {
		case websocket.BinaryMessage:
			go handleAudio(sess, data)
		case websocket.TextMessage:
			go handleControl(sess, data)
		}
	}
}

// writePump writes messages from the send channel to the client.
// It sends periodic pings to keep the connection alive.
// It closes quit and conn on exit to unblock readPump.
func writePump(conn *websocket.Conn, send <-chan any, done <-chan struct{}, quit chan struct{}) {
	defer close(quit)
	defer conn.Close()

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

		send := make(chan any, 8)
		done := make(chan struct{})
		quit := make(chan struct{})

		sess := session.NewSession(send)

		go readPump(conn, sess, done, quit)
		writePump(conn, send, done, quit)
	}
}

func registerWsRoutes(cfg *config.Config, r *gin.Engine) {
	r.GET("/ws", wsHandler(newUpgrader(cfg)))
}
