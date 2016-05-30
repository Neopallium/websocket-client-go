package websocket

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"log"
)

func (s *Socket) sendMsg(msg string) {
	s.out <- []byte(msg)
}

type auxSendEvent struct {
	Event string `json:"event"`
	Data  interface{} `json:"data"`
}

func (s *Socket) SendEvent(event string, data interface{}) {
	e := auxSendEvent{
		Event: event,
		Data: data,
	}
	buf, err := json.Marshal(&e)
	if err != nil {
		log.Fatal("Error sending event:", err)
	}
	s.out <- buf
}

func (s *Socket) makeWriter() {
	ws := s.ws
	out := s.out
	go func () {
		for {
			msg, ok := <-out
			if ! ok {
				// stop writer
				return
			}
log.Println("----------- Send:", string(msg))
			err := ws.WriteMessage(websocket.TextMessage, msg)
			if err != nil {
				log.Println("Writer error:", err)
				return
			}
		}
	} ()
}

