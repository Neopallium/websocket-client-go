package main

import (
  "github.com/Neopallium/pusher-client-go/pusher"
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

func (o *chanHandler) HandleEvent(event pusher.Event) {
  fmt.Println("Handler: Got Event:", event.Channel, event.Event, event.Data)
}

// Example Handler function
func eventHandler(event pusher.Event) {
  fmt.Println("HandlerFunc: Got Event:", event.Channel, event.Event, event.Data)
}

func main() {
  var url, key, channel, event string
  // parse command-line
  flag.StringVar(&url, "url", "", "Pusher url.")

  flag.StringVar(&key, "key", "", "Pusher app key.")

  flag.StringVar(&channel, "channel", "", "Channel subject. (required)")

  flag.StringVar(&event, "event", "", "Channel event to bind to. (optional)")

  flag.Parse()

  var client *pusher.Client

  switch {
  case channel == "":
     errorUsage("Missing required `channel` flag.")
  case key != "" && url != "":
     errorUsage("Can't set both `url` and `key` flags")
  case key != "":
    fmt.Printf("Connect to Pusher app key: %s\n", key)
    client = pusher.NewClient(key)
  case url != "":
    fmt.Printf("Connect to Pusher url: %s\n", url)
    c, err := pusher.NewClientUrl(url)
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
}

