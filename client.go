package main

import (
	"database/sql"

	"github.com/gorilla/websocket"
	_ "github.com/lib/pq"
)

type FindHandler func(string) (Handler, bool)

type Message struct {
	Name string      `json:"name"`
	Data interface{} `json:"data"`
}

type Client struct {
	send         chan Message
	socket       *websocket.Conn
	findhandler  FindHandler
	session      *sql.DB
	stopChannels map[int]chan bool
}

func (c *Client) NewStopChannel(stopKey int) chan bool {
	c.StopForKey(stopKey)
	stop := make(chan bool)
	c.stopChannels[stopKey] = stop
	return stop
}

func (c *Client) StopForKey(key int) {
	if ch, found := c.stopChannels[key]; found {
		ch <- true
		delete(c.stopChannels, key)
	}
}

func (client *Client) Read() {
	var message Message
	for {
		if err := client.socket.ReadJSON(&message); err != nil {
			break
		}
		//what function to call
		if handler, found := client.findhandler(message.Name); found {
			handler(client, message.Data)
		}
	}
	client.socket.Close()
}
func (client *Client) Write() {
	for msg := range client.send {
		if err := client.socket.WriteJSON(msg); err != nil {
			break
		}
	}
	client.socket.Close()
}

func (c *Client) Close() {
	for _, ch := range c.stopChannels {
		ch <- true
	}
	close(c.send)
}

func NewClient(socket *websocket.Conn, findhandler FindHandler, session *sql.DB) *Client {
	return &Client{
		send:         make(chan Message),
		socket:       socket,
		findhandler:  findhandler,
		session:      session,
		stopChannels: make(map[int]chan bool),
	}
}
