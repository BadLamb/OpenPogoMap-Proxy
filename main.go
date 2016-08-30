package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/websocket"
	"net/http"
	"strconv"
)

var exit_hub *Hub

type Request struct {
	Meth string `json:"meth"`
	Host string `json:"host"`
	Cont string `json:"cont"`
	User string `json:"user"`
	Data string `json:"data"`
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func main() {
	log.Info("Started the hub")

	exit_hub = NewHub()
	go exit_hub.Listen()

	http.HandleFunc("/websocket", wsHandler)
	http.HandleFunc("/", requestHandler)
	log.Info("Started the http server")
	http.ListenAndServe(":8080", nil)
}

func requestHandler(w http.ResponseWriter, r *http.Request) {
	//These headers are needed to route the request through the network
	proxy_id := r.Header.Get("Proxy-Id")
	final_host := r.Header.Get("Final-host")

	id, err := strconv.Atoi(proxy_id)
	if err != nil {
		log.Error(err)
		http.Error(w, `Internal error`, http.StatusBadRequest)
		return
	}

	proxy, err := exit_hub.Search(id)
	if err != nil {
		log.Error("Invalid proxy_id!")
		http.Error(w, `Internal Error`, http.StatusBadRequest)
		return
	}

	data := new(bytes.Buffer)
	data.ReadFrom(r.Body)
	data_final := base64.StdEncoding.EncodeToString(data.Bytes())
	req := Request{r.Method, final_host, r.Header.Get("Content-Type"), r.Header.Get("User-Agent"), data_final}

	json, _ := json.Marshal(req)
	proxy.Send(json)

	resp := <-proxy.Response
	w.Write(resp.Data)
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Error("Failed to upgrade:", err)
	}

	defer conn.Close()


	var new_client = NewClient(conn, exit_hub)
	log.Info("New client " + string(new_client.Id))
	exit_hub.Add(new_client)

	new_client.Listen()
}
