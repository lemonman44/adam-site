package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/websocket"
)

func main() {
	fmt.Println("Begin")
	initChatSocketServer()
}

func initChatSocketServer() {
	// make and run the global chat room
	room := &ChatRoom{
		conns:     make(map[*ChatConnection]bool),
		addconn:   make(chan *ChatConnection),
		delconn:   make(chan *ChatConnection),
		broadcast: make(chan ChatMessage),
	}
	go room.run()
	// define paths
	http.HandleFunc("/api/go/sockets/chat", func(w http.ResponseWriter, r *http.Request) {
		chatSocketHandler(room, w, r)
	})
	// host
	err := http.ListenAndServe(":8082", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func chatSocketHandler(c *ChatRoom, w http.ResponseWriter, r *http.Request) {
	fmt.Println("Chat Connection Begin")
	// set some initial websocket settings
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
	// if local then dont check origin
	if os.Getenv("APP_ENV") != "production" {
		upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	}
	// attempt upgrade to websocket protocol
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("Err", err)
		return
	}
	// make ChatConnection wrapper
	chatconn := &ChatConnection{
		room: c,
		conn: conn,
		send: make(chan []byte),
	}
	// add conn to chat room
	c.addconn <- chatconn
	// begin read and write loops
	go chatSocketReadLoop(chatconn)
	go chatSocketWriteLoop(chatconn)
}

// loop to read messages from websocket connections to chat room
func chatSocketReadLoop(c *ChatConnection) {
	// define close process for when this function exits
	defer func() {
		c.room.delconn <- c
		c.conn.Close()
	}()
	// set up a deadline deadline, 3 hours
	realdeadline := time.After(3 * time.Hour)
	// read messages
	for {
		// timeout if nothing has been sent in lets say 1 hour
		deadline := time.Now().Add(1 * time.Hour)
		c.conn.SetReadDeadline(deadline)
		// wait for message
		select {
		case <-realdeadline:
			return
		default:
			msgType, msg, err := c.conn.ReadMessage()
			// if message resulted in error or closing etc then close on this side and exit
			if err != nil {
				fmt.Println("Error: ", err, "M Type: ", msgType)
				return
			}
			// send message to conns in chat room
			fmt.Println("Msg received from client", msg)
			c.room.broadcast <- ChatMessage{msg, len(c.room.conns), c}
		}
	}
}

// loop to write messages from chat room to websocket connections
func chatSocketWriteLoop(c *ChatConnection) {
	// define close process for when this function exits
	defer func() {
		c.room.delconn <- c
		c.conn.Close()
	}()
	for {
		// wait for incoming message from the channel with chat room
		if message, ok := <-c.send; ok {
			fmt.Println("Msg pushing to client", message)
			// write incoming message to
			c.conn.WriteMessage(websocket.TextMessage, message)
			// Add queued chat messages to the current websocket message.
			n := len(c.send)
			for i := 0; i < n; i++ {
				c.conn.WriteMessage(websocket.TextMessage, <-c.send)
			}
		} else {
			// The room closed the send channel.
			c.conn.WriteMessage(websocket.CloseMessage, []byte{})
			return
		}
	}
}

// represents a pool of connections
type ChatRoom struct {
	// map of the connections (mem address as keys)
	conns map[*ChatConnection]bool
	// channel to add connections
	addconn chan *ChatConnection
	// channel to remove connections
	delconn chan *ChatConnection
	// channel for connections to send messages through
	broadcast chan ChatMessage
}

// gives ChatRoom structs a run function to be run asynchronously
func (c *ChatRoom) run() {
	fmt.Println("Chat Room Begin")
	// wait for connections to be added or removed from the room
	go func() {
		for {
			conn := <-c.addconn // if a connection is joining the room, add connection to map
			c.conns[conn] = true
			// send a num conn update
			c.broadcast <- ChatMessage{nil, len(c.conns), nil}
		}
	}()
	go func() {
		for {
			conn := <-c.delconn // if a connection is leaving the room, remove connection from map, and clean up resources
			if _, ok := c.conns[conn]; ok {
				delete(c.conns, conn)
				close(conn.send)
			}
			// send a num conn update
			c.broadcast <- ChatMessage{nil, len(c.conns), nil}
		}
	}()
	go func() {
		for {
			msg := <-c.broadcast // if a message is being sent into the room, send to other connections
			json, err := msg.toJSON()
			if err != nil {
				fmt.Println("Could not convert to JSON: ", msg)
				continue
			}
			for conn := range c.conns {
				if conn == msg.fromconn {
					continue
				}
				conn.send <- json
			}
		}
	}()
}

// wrapper for attaching a channel to a connection
type ChatConnection struct {
	// chat room this connection belongs to
	room *ChatRoom
	// underlying websocket connection
	conn *websocket.Conn
	// channel access into the underlying websocket connection
	send chan []byte
}

type ChatMessage struct {
	message  []byte
	numconns int
	fromconn *ChatConnection
}

func (c ChatMessage) toJSON() ([]byte, error) {
	return json.Marshal(map[string]any{"message": string(c.message), "numconns": c.numconns})
}
