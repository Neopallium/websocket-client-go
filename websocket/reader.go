package websocket

import (
	"github.com/gorilla/websocket"
	"encoding/json"
	"log"
)

func readEvent(ws *websocket.Conn, event *Event) error {
	// 'Data' can be an object or string.  So we use this struct for decoding.
	var aux struct {
		Event string `json:"event"`
		Channel string `json:"channel"`
		Data  interface{} `json:"data"`
	}
	// parse into aux struct.
	if err := ws.ReadJSON(&aux); err != nil {
		return err
	}
log.Println("Got msg:", aux)
	// Copy parsed event.
	event.Event = aux.Event
	event.Channel = aux.Channel
	// make sure "Data" is a string
	switch aux.Data.(type) {
	case string:
		event.Data = aux.Data.(string)
	default:
		buf, err := json.Marshal(aux.Data)
		if err != nil {
			// This shouldn't happen, since we just decoded 'Data' from JSON
			log.Fatal("JSON Marshaller failed:", err)
		}
		event.Data = string(buf)
	}
	return nil
}

// reader goroutine
func (s *Socket) makeReader() {
	ws := s.ws
	in := make(chan *Event, IN_CHANNEL_SIZE)
	go func () {
		for {
			var event Event
			if err := readEvent(ws, &event); err != nil {
				// Close channel to signal that the WebSocket connection has closed.
				close(in)
				return
			}
			in <- &event
		}
	} ()
	s.in = in
}

