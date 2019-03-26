package main

import (
	"flag"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
)

var addr = flag.String("addr", "localhost:8080", "http service address")

var upgrader = websocket.Upgrader{}

func distribute_message(ws *websockets, mt int, ms []byte) {
	for _, websocket := range ws.connections {
		err := websocket.conn.WriteMessage(mt, ms)
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

func (ws *websockets) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	conn := connection{conn: c, name: "unchi"}
	ws.connections = append(ws.connections, conn)
	if err != nil {
		log.Println("upgrade: ", err)
		return
	}
	defer c.Close()

	for {
		mt, message, err := conn.conn.ReadMessage()
		if err != nil {
			log.Println("read: ", err)
			break
		}
		// log.Printf("%s: %s", conn.name, message)
		log.Printf("recv: %s", message)
		distribute_message(ws, mt, message)
	}
}

func main() {
	ws := &websockets{}
	flag.Parse()
	log.SetFlags(0)
	http.Handle("/ws", ws)
	log.Fatal(http.ListenAndServe(*addr, nil))
}
