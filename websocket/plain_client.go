package websocket

import (
	"net/url"
	"time"
)

type PlainClient struct {
	sock         *Socket
}

func (c *PlainClient) HandleDisconnect() bool {
	return true
}

func (c *PlainClient) HandleConnected() {
}

func (c *PlainClient) HandleMessage(msg []byte) error {
	return nil
}

func (c *PlainClient) SendMessage(msg []byte) {
	c.sock.SendMessage(msg)
}

func (c *PlainClient) SendPing() {
	c.sock.SendMessage([]byte("PING"))
}

func (c *PlainClient) Close() {
	c.sock.Close()
}

type Config struct {
	ConnectTimeout    time.Duration
	ActivityTimeout   time.Duration
	PingTimeout       time.Duration
}

var DefaultConfig = Config{
	ConnectTimeout:  time.Second * 30,
	ActivityTimeout: time.Second * 120,
	PingTimeout:     time.Second * 30,
}

func (cf Config) NewClient(websocketUrl string) (Client, error) {
	u, err := url.Parse(websocketUrl)
	if err != nil {
		return nil, err
	}
	p := &PlainClient{}
	p.sock = NewSocket(u, cf, p)
	return p, nil
}

func NewClient(url string) (Client, error) {
	return DefaultConfig.NewClient(url)
}

