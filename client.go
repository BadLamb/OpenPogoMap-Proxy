package main

import (
	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/websocket"
	"time"
)

var maxId int = 0


const (
    // Time allowed to read the next pong message from the peer.
    pongWait = 30 * time.Second

    // Send pings to peer with this period. Must be less than pongWait.
    pingPeriod = (pongWait * 9) / 10

)

type Client struct {
	Id       int
	conn     *websocket.Conn
	Response chan *Message
	Writer   chan []byte
    Hub      *Hub
}

type Message struct {
	Data   []byte
	sender *Client
	time   int64
}

func NewClient(ws *websocket.Conn, h *Hub) *Client {
	maxId += 1
	return &Client{maxId, ws, make(chan *Message), make(chan []byte), h}
}

func (c *Client) write(mt int, payload []byte) error {
    return c.conn.WriteMessage(mt, payload)
}

func (c *Client) Send(data []byte) {
	c.Writer <- data
}

func (c *Client) readHandler(){
    defer func() {
        c.conn.Close()
        c.Hub.Remove(c.Id)
        c.Response <- &Message{[]byte("The client has disconnected"), c, time.Now().Unix()}
    }()

    for {
        _, mes, err := c.conn.ReadMessage()
        if err != nil {
            log.Info("Listener: client disconnected " + string(c.Id))
            break
        }
        c.Response <- &Message{mes, c, time.Now().Unix()}
    }
}

func (c *Client) Listen() {
    ticker := time.NewTicker(pingPeriod)

    go c.readHandler()
	defer func() {
		c.conn.Close()
		c.Hub.Remove(c.Id)
        c.Response <- &Message{[]byte("The client has disconnected"), c, time.Now().Unix()}
	}()

	for {
        select{
        case toWrite := <-c.Writer:
	    	err := c.conn.WriteMessage(1, toWrite)
	    	if err != nil {
	    		log.Info("Failed to write to ws")
                break
		    }

        case <-ticker.C:
            if err := c.write(websocket.PingMessage, []byte{}); err != nil {
                return
            }
	    }
    }
}
