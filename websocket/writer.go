package websocket

import (
	"github.com/gorilla/websocket"
	"log"
)

func (s *Socket) SendMessage(msg Message) error {
	buf, err := msg.EncodeMessage()
	if err != nil {
		return err
	}
	log.Println("------------------ Sock.SendMessage", string(buf))
	s.out <- buf
	return nil
}

func (s *Socket) SendMsg(msg []byte) {
	s.out <- msg
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

