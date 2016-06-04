package pusher

import (
	"encoding/json"
	"log"
)

type Event struct {
	Event    string `json:"event,omitempty"`
	Channel  string `json:"channel,omitempty"`
	Data     interface{} `json:"data,omitempty"`
}

func (e *Event) GetEvent() string {
	return e.Event
}

func (e *Event) SetEvent(event string) {
	e.Event = event
}

func (e *Event) GetChannel() string {
	return e.Channel
}

func (e *Event) SetChannel(channel string) {
	e.Channel = channel
}

func (e *Event) GetData() interface{} {
	return e.Data
}

func (e *Event) SetData(data interface{}) {
	e.Data = data
}

func (e *Event) GetDataString() string {
	// Normalize Data as a string value.
	switch e.Data.(type) {
	case string:
		return e.Data.(string)
	default:
		buf, err := json.Marshal(e.Data)
		if err != nil {
			// This shouldn't happen, since we just decoded 'Data' from JSON
			log.Fatal("JSON Marshaller failed:", err)
		}
		return string(buf)
	}
}

func (e *Event) SetDataString(data string) {
	e.Data = data
}


