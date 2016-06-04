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
}

func (p *PusherClient) handleError(event *Event) error {
	var msg struct {
		Message string
		Code    int64
	}

	data := event.GetData()
	switch data.(type) {
	case string:
		var err error
		err = json.Unmarshal([]byte(data.(string)), &msg)
		if err != nil {
			log.Println("Failed to unmarshal:", event.Event, err)
		}
	default:
		dataMap := data.(map[string]string)
		msg.Message = dataMap["Message"]
		code, err := strconv.ParseInt(dataMap["Code"], 10, 32)
		if err != nil {
			log.Println("Bad 'Code' in error message.")
		}
		msg.Code = code
	}
	switch {
	case 4000 <= msg.Code && msg.Code <= 4099:
		log.Println("Connect failed websocket error: code:", msg.Code, ", message:", msg.Message)
		return ws.ErrClosed
	case 4100 <= msg.Code && msg.Code <= 4199:
		log.Println("Try again (delayed reconnect): code:", msg.Code, ", message:", msg.Message)
		return ws.ErrDelayReconnect
	case 4200 <= msg.Code && msg.Code <= 4299:
		log.Println("Reconnect (no delay): code:", msg.Code, ", message:", msg.Message)
		return ws.ErrReconnect
	default:
		log.Println("Pusher error: code:", msg.Code, ", message:", msg.Message)
	}
	return nil
}

func (p *PusherClient) handleConnectionEstablished(event *Event) error {
	// process Connected information.
	var msg struct {
		SocketId         string `json:"socket_id"`
		ActivityTimeout  int `json:"activity_timeout"`
	}
	if err := json.Unmarshal([]byte(event.GetDataString()), &msg); err != nil {
		log.Println("Failed to unmarshal:", event.Event, err)
	}
	// update activity_timeout value
	p.sock.SetActivityTimeout(time.Duration(msg.ActivityTimeout) * time.Second)
	// subscribe to channels.
	p.channels.ConnectedState(true)
	return nil
}

func (p *PusherClient) HandleMessage(msg []byte) error {
	var err error
	var event Event

	// parse into aux struct.
	if err := json.Unmarshal(msg, &event); err != nil {
		return err
	}
	log.Println("---------- PusherClient event:", event)
	// handle Websocket events.
	switch event.Event {
	case "pusher:ping":
		p.sock.SendMessage([]byte(`{"event":"pusher:pong","data":"{}"}`))
	case "pusher:pong":
		p.sock.HandlePong()
	case "pusher:error":
		err = p.handleError(&event)
	case "pusher:connection_established":
		err = p.handleConnectionEstablished(&event)
	}
	p.channels.HandleEvent(&event)
	return err
}

func (p *PusherClient) SendMessage(msg []byte) {
	log.Println("-------------- PusherClient.SendMessage", string(msg))
	p.sock.SendMessage(msg)
}

func (p *PusherClient) SendPing() {
	// send ping
	p.sock.SendMessage([]byte(`{"event":"pusher:ping","data":"{}"}`))
}

func (p *PusherClient) SendEvent(event string, data interface{}) {
	log.Println("------------------ PusherClient.SendEvent", event, data)
	e := Event{
		Event: event,
		Data: data,
	}
	buf, err := json.Marshal(&e)
	if err != nil {
		log.Fatal("Error sending event:", err)
	}
	p.sock.SendMessage(buf)
}

type subData struct {
	Channel     string `json:"channel"`
	Auth        string `json:"auth,omitempty"`
	ChannelData string `json:"channel_data,omitempty"`
}

func (p *PusherClient) SendSubscribe(channel string) {
	p.SendEvent("pusher:subscribe", &subData{
		Channel: channel,
	})
}

type unsubData struct {
	Channel     string `json:"channel"`
}

func (p *PusherClient) SendUnsubscribe(channel string) {
	p.SendEvent("pusher:unsubscribe", &unsubData{
		Channel: channel,
	})
}

func (p *PusherClient) Close() {
	p.sock.Close()
}

func (p *PusherClient) Subscribe(channel string) ws.Channel {
	ch := p.channels.Find(channel)
	if ch == nil {
		// create a new channel.
		ch = NewPublicChannel(channel, p)
		p.channels.Add(channel, ch)
	}
	return ch
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

