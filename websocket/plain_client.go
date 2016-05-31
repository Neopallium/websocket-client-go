package websocket

import (
	"net/url"
	"time"
	"log"
)

type PlainClient struct {
	sock         *Socket
}

func (c *PlainClient) HandleDisconnect() bool {
	log.Println("------------------ Plain.HandleDisconnect")
	return true
}

func (c *PlainClient) HandleConnected() {
	log.Println("------------------ Plain.HandleConnected")
}

func (c *PlainClient) HandleMessage(msg []byte) error {
	log.Println("------------------ Plain.HandleMessage", string(msg))
	return nil
}

func (c *PlainClient) SendMessage(msg []byte) {
	log.Println("------------------ Plain.SendMessage", string(msg))
	c.sock.SendMessage(msg)
}

func (c *PlainClient) SendPing() {
	log.Println("------------------ Plain.SendPing")
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

