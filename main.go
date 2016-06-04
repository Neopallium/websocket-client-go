package main

import (
  ws "github.com/Neopallium/websocket-client-go/websocket"
  "github.com/Neopallium/websocket-client-go/pusher"
  "flag"
  "fmt"
  "os"
  "os/signal"
  "syscall"
)

func errorUsage(err string) {
  fmt.Fprintln(os.Stderr, err)
  flag.Usage()
  os.Exit(1)
}

// Example Handler object
type chanHandler struct {}

func (o *chanHandler) HandleEvent(event ws.Event) {
  fmt.Println("Handler: Got Event:", event.GetChannel(), event.GetEvent(), event.GetData())
}

// Example Handler function
func eventHandler(event ws.Event) {
  fmt.Println("HandlerFunc: Got Event:", event.GetChannel(), event.GetEvent(), event.GetData())
}

func main() {
  var url, key, channel, event string
  // parse command-line
  flag.StringVar(&url, "url", "", "Pusher url.")

  flag.StringVar(&key, "key", "", "Pusher app key.")

  flag.StringVar(&channel, "channel", "", "Channel subject. (required)")

  flag.StringVar(&event, "event", "", "Channel event to bind to. (optional)")

  flag.Parse()

  var client ws.ChannelClient

  switch {
  case channel == "":
     errorUsage("Missing required `channel` flag.")
  case key != "" && url != "":
     errorUsage("Can't set both `url` and `key` flags")
  case key != "":
    fmt.Printf("Connect to Pusher app key: %s\n", key)
    client = pusher.NewPusher(key)
  case url != "":
    fmt.Printf("Connect to Pusher url: %s\n", url)
    c, err := pusher.NewPusherUrl(url)
    if err != nil {
      errorUsage("Bad url: " + url)
    }
    client = c
  default:
     errorUsage("Can't set both `url` and `key` flags")
  }

  fmt.Println("Subscribe:", channel)
  ch := client.Subscribe(channel)

  fmt.Println("Bind:", event)
  ch.Bind(event, &chanHandler{})
  //ch.BindFunc(event, eventHandler)

  // bind to all channel events
  //client.BindAll(&chanHandler{})
  //client.BindAllFunc(eventHandler)

  fmt.Println("Wait for exit.")
  // block until signal
  s := make(chan os.Signal, 1)
  signal.Notify(s, syscall.SIGHUP)
  signal.Notify(s, syscall.SIGINT)
  <- s

  // close client.
  client.Close()
}

