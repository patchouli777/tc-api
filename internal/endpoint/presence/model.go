package presence

import "golang.org/x/net/websocket"

type Channel string

type InitMessage struct {
	Channel string `json:"channel"`
}

type Conn struct {
	Channel Channel
	Conn    *websocket.Conn
}
