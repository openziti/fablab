package console

import (
	"github.com/sirupsen/logrus"
	"golang.org/x/net/websocket"
)

func NewClient(ws *websocket.Conn, server *Server) *Client {
	nextId++
	return &Client{
		id:     nextId,
		ws:     ws,
		server: server,
		ch:     make(chan *Message, chBufSize),
		doneCh: make(chan struct{}),
	}
}

func (client *Client) Conn() *websocket.Conn {
	return client.ws
}

func (client *Client) Write(msg *Message) {
	select {
	case client.ch <- msg:
	default:
	}
}

func (client *Client) Listen() {
	client.listenWrite()
}

func (client *Client) listenWrite() {
	for {
		select {
		case msg := <-client.ch:
			if err := websocket.JSON.Send(client.ws, msg); err != nil {
				logrus.Errorf("error sending to client [#%d] (%w)", client.id, err)
			}

		case <-client.doneCh:
			client.server.Del(client)
			close(client.doneCh)
			return
		}
	}
}

type Client struct {
	id     int
	ws     *websocket.Conn
	server *Server
	ch     chan *Message
	doneCh chan struct{}
}

var nextId int = 0

const chBufSize = 100
