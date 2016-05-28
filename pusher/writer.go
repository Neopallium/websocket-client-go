package pusher

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"log"
)

func (c *Client) sendMsg(msg string) {
	c.out <- []byte(msg)
}

type auxSendEvent struct {
	Event string `json:"event"`
	Data  interface{} `json:"data"`
}

func (c *Client) sendEvent(event string, data interface{}) {
	e := auxSendEvent{
		Event: event,
		Data: data,
	}
	buf, err := json.Marshal(&e)
	if err != nil {
		log.Fatal("Error sending event:", err)
	}
	c.out <- buf
}

func (c *Client) makeWriter() {
	go func (ws *websocket.Conn, out chan []byte) {
		for {
			msg, ok := <- out
			if ! ok {
				// stop writer
				return
			}
			err := ws.WriteMessage(websocket.TextMessage, []byte(msg))
			if err != nil {
				log.Println("Writer error:", err)
				return
			}
		}
	} (c.ws, c.out)
}

