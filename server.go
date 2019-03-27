package main

import (
	"./socket"
	"flag"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
)

var addr = flag.String("addr", "localhost:8080", "http service address")

var upgrader = websocket.Upgrader{}

func distribute_message(ws *websockets, m socket.Message) {
	for _, websocket := range ws.connections {
		err := websocket.conn.WriteJSON(m)
		if err != nil {
			log.Println("write error: ", err)
			continue
		}
	}
}

type websockets struct {
	connections []connection
}
type connection struct {
	conn *websocket.Conn
	name string
}

func (ws *websockets) json(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	conn := connection{conn: c, name: "unchi"}
	ws.connections = append(ws.connections, conn)
	if err != nil {
		log.Println("upgrade: ", err)
		return
	}
	defer c.Close()

	var m socket.Message
	for {
		err := conn.conn.ReadJSON(&m)
		if err != nil {
			log.Println("read: ", err)
			break
		}
		distribute_message(ws, m)
		log.Printf("recv: %s %s", m.Name, m.Message)
	}
}

func main() {
	ws := &websockets{}
	flag.Parse()
	log.SetFlags(0)
	http.HandleFunc("/ws", ws.json)
	log.Fatal(http.ListenAndServe(*addr, nil))
}
