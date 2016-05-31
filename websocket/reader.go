package websocket

import (
	"log"
)

// reader goroutine
func (s *Socket) makeReader() {
	ws := s.ws
	in := make(chan *Event, IN_CHANNEL_SIZE)
	go func () {
		for {
			_, buf, err := ws.ReadMessage()
			if err != nil {
				// Close channel to signal that the WebSocket connection has closed.
				close(in)
				return
			}
			var event Event
			err = event.DecodeMessage(buf)
			if err != nil {
				log.Println("Failed to decode message:", err)
				continue
			}
			in <- &event
		}
	} ()
	s.in = in
}

