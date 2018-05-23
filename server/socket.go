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

// Conn .
type Conn struct {
	conn *websocket.Conn
	send chan interface{}
}

var wsConns = map[*Conn]bool{}
var wsHandlers = map[string]func(map[string]interface{}, *Conn){}

var wsHandler = func() func(c *gin.Context) {
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
		safeConn := &Conn{conn, make(chan interface{})}
		wsConns[safeConn] = true
		go safeConn.writePump()
		go safeConn.readPump()
	}
}()

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second
	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second
	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10
	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

func (c *Conn) readPump() {
	defer c.conn.Close()
	defer delete(wsConns, c)

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Println(err)
			}
			break
		}
		message = bytes.TrimSpace(message)

		sendErr := func(err error) {
			c.send <- map[string]interface{}{
				"ok":     false,
				"code":   400,
				"reason": err.Error(),
			}
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
		handler(json, c)
	}
}

func (c *Conn) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.conn.WriteJSON(message); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
