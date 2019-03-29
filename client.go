package main

import (
	"./socket"

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
	done := make(chan struct{})
	msg := make(chan string)
	recv_msg := make(chan socket.Message)
	client := socket.InitClient()
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	flag.Parse()
	log.SetFlags(0)

	u := url.URL{Scheme: "ws", Host: *addr, Path: "/ws"}
	log.Printf("connection to %s", u.String())

	// 上記のurl uに接続
	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial: ", err)
	}

	defer fmt.Println("ALL EXIT")
	defer c.Close()

	// コマンド入力
	go input(done, msg)
	// serverからのメッセージ取得
	go read_message(c, recv_msg)

	for {
		select {
		case <-done:
			close(msg)
			return
		// message取得
		case m := <-msg:
			client.Message = string(m)
			err := c.WriteJSON(client)
			if err != nil {
				log.Println("read error: ", err)
				return
			}
		case m := <-recv_msg:
			log.Printf("%s: %s", m.Name, m.Message)
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

func read_message(c *websocket.Conn, recv_msg chan<- socket.Message) {
	var msg socket.Message
	for {
		err := c.ReadJSON(&msg)
		if err != nil {
			log.Println("Error: ", err)
			break
		}
		recv_msg <- msg
	}
}
