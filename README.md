# Reconnecting Websocket Client Go Library

A reconnecting Websocket client.

## Installing

```
$ go get github.com/Neopallium/websocket-client-go
```

## Quick start

You can quickly try the library by using the included simple client.

```
$ websocket-client-go -url URL -channel TEST_CHANNEL -event EVENT_NAME
```

## Getting Started

```go
package main

import (
  "github.com/Neopallium/websocket-client-go/websocket"
  "sync"
  "fmt"
)

var wg sync.WaitGroup

// Example Handler object
type chanHandler struct {}

func (o *chanHandler) HandleEvent(event websocket.Event) {
  fmt.Println("Handler: Got Event:", event.Channel, event.Event, event.Data)
  wg.Done()
}

// Example Handler function
func eventHandler(event websocket.Event) {
  fmt.Println("HandlerFunc: Got Event:", event.Channel, event.Event, event.Data)
  wg.Done()
}

func main() {
  // Create websocket client connection for "url"
  client := websocket.NewClient("wss://localhost:8080/app")

  // Subscribe to a channel
  ch := client.Subscribe("test_channel")

  wg.Add(10) // wait for 10 events.

  // Bind to events on the channel.  Can use either a Handler object.
  ch.Bind("my_event", &chanHandler{})
  // or Handler callback function.
  ch.BindFunc("my_event", eventHandler)

  // bind to all channel events
  client.BindAll(&chanHandler{})
  client.BindAllFunc(eventHandler)

  // wait for some events.
  wg.Wait()
}

```

