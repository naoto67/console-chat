package main

import (
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

	u := url.URL{Scheme: "ws", Host: *addr, Path: "/echo"}
	log.Printf("connection to %s", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial: ", err)
	}

	defer c.Close()

	done := make(chan struct{})
	msg := make(chan string)
	scanner := bufio.NewScanner(os.Stdin)

	go func() {
		defer close(done)
		for {
			scanner.Scan()
			if err := scanner.Err(); err != nil {
				fmt.Println("Scanner Error: ", err)
				return
			}
			msg <- scanner.Text()
		}
	}()

	for {
		select {
		case <-done:
			close(msg)
			return
		case m := <-msg:
			err := c.WriteMessage(websocket.TextMessage, []byte(m))
			if err != nil {
				log.Println("read: ", err)
				return
			}
			log.Printf("recv: %s", m)
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
