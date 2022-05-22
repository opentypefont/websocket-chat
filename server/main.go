package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/bwmarrin/snowflake"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"log"
	"os"
)

type Client struct {
	name string
	id   snowflake.ID
}

var connections = make(map[*websocket.Conn]Client)
var sNode, _ = snowflake.NewNode(1)

func main() {
	port := flag.Int("port", 8000, "port of server")
	flag.Parse()

	app := fiber.New(fiber.Config{DisableStartupMessage: true})

	app.Use(func(ctx *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(ctx) {
			return ctx.Next()
		}

		return ctx.SendStatus(fiber.StatusUpgradeRequired)
	})
	app.Get("/", websocket.New(func(conn *websocket.Conn) {
		defer disconnectHandler(conn)

		client := connectHandler(conn)

		for {
			messageType, message, err := conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					log.Println("read error:", err)
				}

				return
			}

			if messageType == websocket.TextMessage {
				broadcastMessage(fmt.Sprintf("%s: %s", client.name, message))
			}
		}
	}))

	go serverMessageSender()

	log.Printf("listening on port %d", *port)
	if err := app.Listen(fmt.Sprintf(":%d", *port)); err != nil {
		log.Fatalln(err)
	}
}

func serverMessageSender() {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		broadcastMessage("Server: " + scanner.Text())
	}
}

func connectHandler(conn *websocket.Conn) Client {
	id := sNode.Generate()
	client := Client{
		name: "Guest " + id.String()[10:],
		id:   id,
	}
	connections[conn] = client

	broadcastMessage(client.name + " has connected")

	return client
}

func disconnectHandler(conn *websocket.Conn) {
	broadcastMessage(connections[conn].name + " has disconnected")

	delete(connections, conn)
	err := conn.Close()
	if err != nil {
		log.Println("close error: ", err)
	}
}

func broadcastMessage(message string) {
	for conn := range connections {
		err := conn.WriteMessage(websocket.TextMessage, []byte(message))
		if err != nil {
			log.Println("write error: ", err)
		}
	}
	log.Println(message)
}
