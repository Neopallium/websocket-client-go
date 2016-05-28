package pusher

import (
	"net/url"
	"time"
)

type Pusher struct {
	Client            string
	Version           string
	Protocol          int
	ConnectTimeout    time.Duration
	ActivityTimeout   time.Duration
	PingTimeout       time.Duration
}

var DefaultPusher = &Pusher{
	Client:          "pusher-websocket-go",
	Version:         "0.5",
	Protocol:        7,
	ConnectTimeout:  time.Second * 30,
	ActivityTimeout: time.Second * 120,
	PingTimeout:     time.Second * 30,
}

func (p *Pusher) NewClientUrl(pusherUrl string) (*Client, error) {
	u, err := url.Parse(pusherUrl)
	if err != nil {
		return nil, err
	}
	return newClient(u, p), nil
}

func (p *Pusher) NewClient(appKey string) (*Client) {
	u := &url.URL{
		Scheme: "wss",
		Host: "ws.pusherapp.com:443",
		Path: "/app/" + appKey,
	}
	return newClient(u, p)
}

func NewClientUrl(url string) (*Client, error) {
	return DefaultPusher.NewClientUrl(url)
}

func NewClient(appKey string) (*Client) {
	return DefaultPusher.NewClient(appKey)
}

