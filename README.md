# Pusher Websocket Client Go Library

Golang library for connecting to a Pusher App over websockets.

## Installing

```
$ go get github.com/Neopallium/pusher-client-go
```

## Quick start

You can quickly try the library by using the included simple client.

```
$ pusher-client-go -key APP_KEY -channel TEST_CHANNEL -event EVENT_NAME
```

## Getting Started

```go
package main

import (
  "github.com/Neopallium/pusher-client-go/pusher"
  "sync"
  "fmt"
)

var wg sync.WaitGroup

// Example Handler object
type chanHandler struct {}

func (o *chanHandler) HandleEvent(event pusher.Event) {
  fmt.Println("Handler: Got Event:", event.Channel, event.Event, event.Data)
  wg.Done()
}

// Example Handler function
func eventHandler(event pusher.Event) {
  fmt.Println("HandlerFunc: Got Event:", event.Channel, event.Event, event.Data)
  wg.Done()
}

func main() {
  // Create pusher client connection for "app_key"
  client := pusher.NewClient("app_key")

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

## TODO

* Support private & presence channels.  Need auth handler/callback.

