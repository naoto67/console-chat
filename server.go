package main

import (
	"./socket"

	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/rs/xid"
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
	id   string
}

func (ws *websockets) json(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	guid := xid.New()
	conn := connection{conn: c, id: guid.String()}
	fmt.Println(conn)
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
			// wsのconnectionsから削除
			ws.remove(conn.id)
			log.Println("read: ", err)
			// log.Println(len(ws.connections))
			break
		}
		conn.name = m.Name
		distribute_message(ws, m)
		log.Printf("%s: %s", m.Name, m.Message)
	}
}

func (ws *websockets) remove(id string) {
	var index int
	for i, v := range ws.connections {
		if v.id == id {
			index = i
			break
		}
	}
	ws.connections = append(ws.connections[:index], ws.connections[(index+1):]...)
}

func main() {
	ws := &websockets{}
	flag.Parse()
	log.SetFlags(0)
	http.HandleFunc("/ws", ws.json)
	log.Fatal(http.ListenAndServe(*addr, nil))
}
