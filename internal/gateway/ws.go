package gateway

import (
	"encoding/binary"
	"encoding/json"
	"log"
	"math"
	"net/http"
	"strings"
	"time"

	"github.com/alvisLu/go-shorten/internal/config"
	"github.com/alvisLu/go-shorten/internal/session"
	"github.com/alvisLu/go-shorten/internal/stt"
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
	Type          string `json:"type"`
	SourceLang    string `json:"sourceLang"`
	TargetLang    string `json:"targetLang"`
	SampleRate    int    `json:"sampleRate"`
	EnableDenoise bool   `json:"enableDenoise"`
}

func handleControl(sess *session.Session, raw []byte) {
	var msg controlMsg
	if err := json.Unmarshal(raw, &msg); err != nil {
		sess.Send(session.WsErrorResp{Error: "invalid JSON: " + err.Error()})
		return
	}
	switch msg.Type {
	case "start":
		sess.Start(msg.SourceLang, msg.TargetLang, msg.SampleRate, msg.EnableDenoise)
		sess.Send(session.WsResp{Status: "started"})
	case "stop":
		sess.Stop()
		sess.Send(session.WsResp{Status: "stopped"})
	case "health":
		sess.Health()
	default:
		sess.Send(session.WsErrorResp{Error: "unknown control type"})
	}
}

func handleAudio(sess *session.Session, pipeline *stt.Pipeline, data []byte) {
	const headerLen = 23
	if !sess.IsRunning() || len(data) < headerLen {
		return
	}

	isFinal := data[0] == 1
	var chName string
	switch data[1] {
	case 0:
		chName = "mic"
	case 1:
		chName = "loopback"
	default:
		return
	}
	id := strings.TrimRight(string(data[2:23]), "\x00")

	pcmRaw := data[23:]
	if len(pcmRaw)%4 != 0 {
		return
	}
	pcm := make([]float32, len(pcmRaw)/4)
	for i := range pcm {
		bits := binary.LittleEndian.Uint32(pcmRaw[i*4:])
		pcm[i] = math.Float32frombits(bits)
	}

	if isFinal {
		pipeline.OnFinalFrame(sess, chName, id, pcm)
	} else {
		pipeline.OnInterimFrame(sess, chName, id, pcm)
	}
}

// readPump reads messages from the client and dispatches them.
// It also handles pong messages to reset the read deadline.
func readPump(conn *websocket.Conn, sess *session.Session, pipeline *stt.Pipeline, done chan<- struct{}, quit <-chan struct{}) {
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
			handleAudio(sess, pipeline, data)
		case websocket.TextMessage:
			log.Printf("ws readPump: %s", data)
			handleControl(sess, data)
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
			if raw, ok := msg.([]byte); ok {
				if err := conn.WriteMessage(websocket.BinaryMessage, raw); err != nil {
					return
				}
			} else {
				if err := conn.WriteJSON(msg); err != nil {
					return
				}
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

func wsHandler(upgrader websocket.Upgrader, pipeline *stt.Pipeline) gin.HandlerFunc {
	return func(c *gin.Context) {
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		log.Printf("ws connected: %s", conn.RemoteAddr())

		send := make(chan any, 8)
		done := make(chan struct{})
		quit := make(chan struct{})

		sess := session.NewSession(send)

		go readPump(conn, sess, pipeline, done, quit)
		writePump(conn, send, done, quit)
	}
}

func registerWsRoutes(cfg *config.Config, r *gin.Engine, pipeline *stt.Pipeline) {
	r.GET("/ws", wsHandler(newUpgrader(cfg), pipeline))
}
