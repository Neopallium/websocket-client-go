package websocket

import (
	"encoding/json"
	"log"
)

type Event struct {
	Event    string `json:"event,omitempty"`
	Channel  string `json:"channel,omitempty"`
	Data     string `json:"data,omitempty"`
}

func (e *Event) DecodeMessage(buf []byte) error {
	// 'Data' can be an object or string.  So we use this struct for decoding.
	var aux struct {
		Event string `json:"event"`
		Channel string `json:"channel"`
		Data  interface{} `json:"data"`
	}
	// parse into aux struct.
	if err := json.Unmarshal(buf, &aux); err != nil {
		return err
	}
log.Println("Got msg:", aux)
	// Copy parsed event.
	e.Event = aux.Event
	e.Channel = aux.Channel
	// make sure "Data" is a string
	switch aux.Data.(type) {
	case string:
		e.Data = aux.Data.(string)
	default:
		buf, err := json.Marshal(aux.Data)
		if err != nil {
			// This shouldn't happen, since we just decoded 'Data' from JSON
			log.Fatal("JSON Marshaller failed:", err)
		}
		e.Data = string(buf)
	}
	return nil
}

func (e *Event) EncodeMessage() ([]byte, error) {
	return nil, nil
}

