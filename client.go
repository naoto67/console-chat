package main

import (
	"./message"
	"bufio"
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/websocket"
)

var addr = flag.String("addr", "localhost:8080", "http service address")

func main() {
	flag.Parse()
	log.SetFlags(0)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	u := url.URL{Scheme: "ws", Host: *addr, Path: "/ws"}
	log.Printf("connection to %s", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial: ", err)
	}

	defer fmt.Println("ALL EXIT")
	defer c.Close()

	done := make(chan struct{})
	msg := make(chan string)

	go input(done, msg)

	recv_msg := make(chan message.Message)
	go read_message(c, recv_msg)

	for {
		select {
		case <-done:
			close(msg)
			return
		case m := <-msg:
			// err := c.WriteMessage(websocket.TextMessage, []byte(m))
			msg := message.Message{
				Name:    "sample",
				Message: string(m),
			}
			err := c.WriteJSON(msg)
			if err != nil {
				log.Println("read: ", err)
				return
			}
		case m := <-recv_msg:
			fmt.Println(m)
		// プロセスを直接切った時などに入る
		case <-interrupt:
			log.Println("interrupt")
			err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Println("write close: ", err)
				return
			}

			select {
			case <-done:
			case <-time.After(time.Second):
			}
			return
		}
	}
}

func input(done chan<- struct{}, msg chan<- string) {
	scanner := bufio.NewScanner(os.Stdin)
	defer close(done)
	for {
		scanner.Scan()
		if err := scanner.Err(); err != nil {
			fmt.Println("Scanner Error: ", err)
			break
		}
		msg <- scanner.Text()
	}
}

func read_message(c *websocket.Conn, recv_msg chan<- message.Message) {
	var msg message.Message
	for {
		err := c.ReadJSON(&msg)
		if err != nil {
			log.Println("Error: ", err)
			break
		}
		recv_msg <- msg
	}
}
