package server

import (
	"bytes"
	"errors"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/zhengxiaoyao0716/util/easyjson"
)

var wsConns = map[*websocket.Conn]bool{}
var wsHandlers = map[string]func(map[string]interface{}, *websocket.Conn){}

var wsHandler = func() func(c *gin.Context) {
	const (
		// Time allowed to write a message to the peer.
		writeWait = 10 * time.Second
		// Time allowed to read the next pong message from the peer.
		pongWait = 60 * time.Second
		// Maximum message size allowed from peer.
		maxMessageSize = 512
	)

	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
	return func(c *gin.Context) {
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			log.Println(err)
			return
		}
		defer conn.Close()

		wsConns[conn] = true
		defer delete(wsConns, conn)

		conn.SetReadLimit(maxMessageSize)
		conn.SetReadDeadline(time.Now().Add(pongWait))

		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					log.Printf("error: %v", err)
				}
				break
			}
			message = bytes.TrimSpace(message)

			sendErr := func(err error) {
				conn.WriteJSON(map[string]interface{}{
					"ok":     false,
					"code":   400,
					"reason": err.Error(),
				})
			}
			data, err := easyjson.Parse(string(message))
			if err != nil {
				sendErr(err)
				continue
			}
			json, err := easyjson.ObjectOf(data)
			if err != nil {
				sendErr(err)
				continue
			}
			action, err := json.StringAt("action")
			if err != nil {
				sendErr(err)
				continue
			}
			delete(json, "action")
			handler, ok := wsHandlers[string(action)]
			if !ok {
				sendErr(errors.New("no handler found for action: " + string(action)))
				continue
			}
			handler(json, conn)
		}
	}
}()
