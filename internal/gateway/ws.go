package gateway

import (
	"encoding/json"
	"net/http"

	"github.com/alvisLu/go-shorten/internal/ws"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type wsDispatcher map[string]func() (any, error)

func newWsDispatcher(svc ws.WsService) wsDispatcher {
	return wsDispatcher{
		"health": func() (any, error) { return svc.WsHealth(), nil },
	}
}

func wsHandler(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	defer conn.Close()

	dispatch := newWsDispatcher(ws.NewService())

	for {
		_, data, err := conn.ReadMessage()
		if err != nil {
			break
		}

		var msg ws.WsReq
		if err := json.Unmarshal(data, &msg); err != nil {
			break
		}

		if handler, ok := dispatch[msg.Type]; ok {
			resp, err := handler()
			if err != nil {
				break
			}
			if err := conn.WriteJSON(resp); err != nil {
				break
			}
		}
	}
}

func registerWsRoutes(r *gin.Engine) {
	r.GET("/ws", wsHandler)
}
