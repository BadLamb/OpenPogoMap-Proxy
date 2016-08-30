package main

import (
	"errors"
)

type Hub struct {
	proxies     map[int]*Client
	RegisterC   chan *Client
	UnRegisterC chan int
}

func NewHub() *Hub {
	proxies := make(map[int]*Client)
	register := make(chan *Client)
	unregister := make(chan int)
	return &Hub{proxies, register, unregister}
}

func (h *Hub) Add(c *Client) {
	h.RegisterC <- c
}

func (h *Hub) Remove(proxy_id int) {
	h.UnRegisterC <- proxy_id
}

func (h *Hub) Search(proxy_id int) (*Client, error) {
	if val, ok := h.proxies[proxy_id]; ok {
		return val, nil
	}
	return nil, errors.New("Proxy not found")
}

func (h *Hub) Listen() {
	for {
		select {

		case client := <-h.RegisterC:
			h.proxies[client.Id] = client

		case client := <-h.UnRegisterC:
			delete(h.proxies, client)

		}
	}
}
