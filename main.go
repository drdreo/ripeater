package main

import (
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"

	"github.com/jaevor/go-nanoid"
)

type Client struct {
	Id         string
	Connection *websocket.Conn
}

type ClientConnections map[string]Client

func getPort() string {
	port := os.Getenv("PORT")
	if port == "" {
		port = ":3000"
	} else {
		port = ":" + port
	}

	return port
}

func connectHandler(connections *ClientConnections, connection *websocket.Conn) Client {
	canonicID, _ := nanoid.Standard(21)
	connId := canonicID()
	log.Printf("New ws request %s", connId)
	client := Client{Id: connId, Connection: connection}
	(*connections)[connId] = client
	return client
}
func disconnectHandler(connections *ClientConnections, client Client) {
	log.Printf("disconnectHandler %s", client.Id)
	delete((*connections), client.Id)
}

func main() {

	clientConnections := make(ClientConnections)
	app := fiber.New()

	app.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message": "Hello, from Ripeater!",
		})
	})

	// Access the websocket server: ws://localhost:3000/ws
	app.Get("/ws", websocket.New(func(c *websocket.Conn) {
		client := connectHandler(&clientConnections, c)

		c.SetCloseHandler(func(code int, text string) error {
			disconnectHandler(&clientConnections, client)
			return nil
		})

		for {
			mtype, msg, err := c.ReadMessage()
			if err != nil {
				if websocket.IsCloseError(err) {
					log.Println("Disonnected")
				}
				log.Println("Read:", err)
				break
			}

			log.Printf("Incoming: %s", msg)

			for _, cc := range clientConnections {
				err = cc.Connection.WriteMessage(mtype, msg)
				if err != nil {
					log.Println("Write:", err)
					break
				}
			}
		}
	}))

	log.Fatal(app.Listen(getPort()))
}
