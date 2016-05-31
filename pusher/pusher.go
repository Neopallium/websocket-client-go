package pusher

import (
	ws "github.com/Neopallium/websocket-client-go/websocket"

	"encoding/json"
	"net/url"
	"time"
	"strconv"
	"log"
)

type PusherClient struct {
	sock               *ws.Socket
	channels           *ws.Channels
}

func (p *PusherClient) HandleDisconnect() bool {
	log.Println("---------- PusherClient Disconnected:")
	p.channels.ConnectedState(false)
	return true
}

func (p *PusherClient) HandleConnected() {
	log.Println("---------- PusherClient Connected:")
	p.channels.ConnectedState(true)
}

func (p *PusherClient) DecodeMessage(buf []byte) (ws.Message, error) {
	var event ws.Event

	err := event.DecodeMessage(buf)
	if err != nil {
		return nil, err
	}
	return &event, nil
}

func (p *PusherClient) HandleMessage(msg ws.Message) ws.ChangeState {
	event := msg.(*ws.Event)
	log.Println("---------- PusherClient event:", event)
	p.channels.HandleEvent(*event)
	return ws.NoChangeState
}

func (p *PusherClient) SendMessage(msg ws.Message) {
	log.Println("-------------- PusherClient.SendMessage", msg)
	p.sock.SendMessage(msg)
}

type auxSendEvent struct {
	Event string `json:"event"`
	Data  interface{} `json:"data"`
}

func (p *PusherClient) SendEvent(event string, data interface{}) {
	log.Println("------------------ PusherClient.SendEvent", event, data)
	e := auxSendEvent{
		Event: event,
		Data: data,
	}
	buf, err := json.Marshal(&e)
	if err != nil {
		log.Fatal("Error sending event:", err)
	}
	p.sock.SendMsg(buf)
}

func (p *PusherClient) Close() {
	p.sock.Close()
}

func (p *PusherClient) Subscribe(channel string) *ws.Channel {
	return p.channels.Add(channel)
}

func (p *PusherClient) Unsubscribe(channel string) {
	p.channels.Remove(channel)
}

func (p *PusherClient) Bind(event string, h ws.Handler) {
	p.channels.Bind(event, h)
}

func (p *PusherClient) Unbind(event string, h ws.Handler) {
	p.channels.Unbind(event, h)
}

func (p *PusherClient) BindFunc(event string, h func(ws.Event)) {
	p.Bind(event, ws.HandlerFunc(h))
}

func (p *PusherClient) UnbindFunc(event string, h func(ws.Event)) {
	p.Unbind(event, ws.HandlerFunc(h))
}

func (p *PusherClient) BindAll(h ws.Handler) {
	p.Bind("", h)
}

func (p *PusherClient) UnbindAll(h ws.Handler) {
	p.Unbind("", h)
}

func (p *PusherClient) BindAllFunc(h func(ws.Event)) {
	p.Bind("", ws.HandlerFunc(h))
}

func (p *PusherClient) UnbindAllFunc(h func(ws.Event)) {
	p.Unbind("", ws.HandlerFunc(h))
}

func newPusherClient(u *url.URL, cf PusherConfig) *PusherClient {
	params := u.Query()
	params.Set("protocol", strconv.Itoa(cf.Protocol))
	params.Set("version", cf.Version)
	params.Set("client", cf.Client)
	u.RawQuery = params.Encode()
	p := &PusherClient{}
	p.sock = ws.NewSocket(u, cf.Config, p)
	p.channels = ws.NewChannels(p)
	return p
}

type PusherConfig struct {
	ws.Config
	Client            string
	Version           string
	Protocol          int
}

var (
	DefaultPusher = PusherConfig{
		Config: ws.Config{
			ConnectTimeout:  time.Second * 30,
			ActivityTimeout: time.Second * 120,
			PingTimeout:     time.Second * 30,
		},
		Client:          "pusher-websocket-go",
		Version:         "0.5",
		Protocol:        7,
	}
)

func (p PusherConfig) NewPusherUrl(pusherUrl string) (*PusherClient, error) {
	u, err := url.Parse(pusherUrl)
	if err != nil {
		return nil, err
	}
	return newPusherClient(u, p), nil
}

func (p PusherConfig) NewPusher(appKey string) (*PusherClient) {
	u := &url.URL{
		Scheme: "wss",
		Host: "ws.pusherapp.com:443",
		Path: "/app/" + appKey,
	}
	return newPusherClient(u, p)
}

func NewPusherUrl(url string) (*PusherClient, error) {
	return DefaultPusher.NewPusherUrl(url)
}

func NewPusher(appKey string) *PusherClient {
	return DefaultPusher.NewPusher(appKey)
}

