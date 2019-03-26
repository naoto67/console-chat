package main

import (
	"flag"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
)

var addr = flag.String("addr", "localhost:8080", "http service address")

var upgrader = websocket.Upgrader{}

func echo(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("upgrade: ", err)
		return
	}
	defer c.Close()
	for {
		mt, message, err := c.ReadMessage()
		if err != nil {
			log.Println("read: ", err)
			break
		}
		log.Printf("recv: %s", message)
		err = c.WriteMessage(mt, message)
		if err != nil {
			log.Println("write: ", err)
			break
		}
	}
}

type AnyHandler struct {
	array []string
}

func (a *AnyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	a.array = append(a.array, "aaa")
	log.Println(a)
	log.Printf("%T", a)
}

func main() {
	anyHandler := &AnyHandler{}
	flag.Parse()
	log.SetFlags(0)
	http.HandleFunc("/echo", echo)
	http.Handle("/sample", anyHandler)
	log.Fatal(http.ListenAndServe(*addr, nil))
}
