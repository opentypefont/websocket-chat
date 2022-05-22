package main

import (
	"bufio"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"os"
	"strings"
)

func main() {
	address := os.Args[1]
	conn, _, err := websocket.DefaultDialer.Dial(fmt.Sprintf("ws://%s/", strings.ReplaceAll(
		strings.ReplaceAll(address, "\n", ""), "\r", "")), nil)
	if err != nil {
		log.Fatalln(err)
	}

	go readHandler(conn)

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		err := conn.WriteMessage(websocket.TextMessage, scanner.Bytes())
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Fatalln(err)
			}
		}
	}
}

func readHandler(conn *websocket.Conn) {
	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Fatalln(err)
			}
		}

		if messageType == websocket.TextMessage {
			log.Println(string(message))
		}
	}
}
