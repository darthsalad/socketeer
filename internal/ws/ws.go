// Internal package for handling websocket connections
// and dispatching updates to clients.
// 
// This package is used in the following way:
// 
// 	1. Create a new WebSocket type with NewWebSocket().
// 	2. Start the WebSocket with Start().
// 	3. Stop the WebSocket with Stop().
//	4. Dispatch updates to clients with DispatchUpdate().
//
// No need to call these methods exclusively, they are
// automatically called and are executed synchronously
// in the socketeer.go file.
package ws

import (
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

// WebSocket is an interface for handling websocket connections.
//
// 	- clients is a map of websocket connections.
// 	- clientsMux is a mutex for clients for thread safety.
type WebSocket struct {
	clients    map[*websocket.Conn]struct{}
	clientsMux sync.Mutex
}

// NewWebSocket returns a new WebSocket.
//
// This method is utilized to create a new WebSocket type 
// and the clients map is initialized which is initially empty.
//
// # Example:
//
// 	conn := ws.NewWebSocket()
func NewWebSocket() *WebSocket {
	return &WebSocket{
		clients: make(map[*websocket.Conn]struct{}),
	}
}

// Start starts the https server and calls the
// websocketHandler method when a connection is made
// to upgrade the connection to a websocket connection.
//
// This method is called internally when the socketeer is started.
//
// # Parameters:
// 
// 	- host (string): the host address to listen on, example: localhost:8080 
// 	- endpoint (string): the endpoint to listen on (without the trailing slash), 
// 		example: /listen 
//
// # Example:
//
// 	ws.Start("localhost:8080", "/listen") // listens on 'ws://localhost:8080/listen' endpoint
func (w *WebSocket) Start(host string, endpoint string) {
	http.HandleFunc(endpoint, w.websocketHandler)
	err := http.ListenAndServe(host, nil)
	if err != nil {
		log.Fatal(err)
	}
}

// Stop stops the websocket server and closes all
// websocket connections.
//
// This method is called internally when the socketeer is stopped.
//
// # Example:
//
// 	ws.Stop()
func (w *WebSocket) Stop() {
	w.clientsMux.Lock()
	defer w.clientsMux.Unlock()

	for client := range w.clients {
		client.Close()
	}

	w.clients = make(map[*websocket.Conn]struct{})
}

// DispatchUpdate dispatches an update to all clients as a
// websocket message in the form of a byte slice.
//
// This method is called internally when an update is received
// from the database.
//
// # Parameters:
//
// 	- update ([]byte): the update to dispatch to clients.
//
// # Example:
//
// 	ws.DispatchUpdate([]byte("Hello, world!"))
func (w *WebSocket) DispatchUpdate(update []byte) {
	w.clientsMux.Lock()
	defer w.clientsMux.Unlock()

	for client := range w.clients {
		err := client.WriteMessage(websocket.TextMessage, update)
		if err != nil {
			log.Println(err)
			return
		}
	}
}

// websocketHandler upgrades the connection to a websocket connection
// and adds the connection to the clients map.
//
// This method is called internally when a connection is made to the
// websocket server.
//
// # Parameters:
//
// 	- res (http.ResponseWriter): the response writer.
// 	- req (*http.Request): the request.
//
// # Example:
//
// 	http.HandleFunc("/listen", ws.websocketHandler)
func (w *WebSocket) websocketHandler(res http.ResponseWriter, req *http.Request) {
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool { 
			return true 
		},
	}
	conn, err := upgrader.Upgrade(res, req, nil)
	if err != nil {
		log.Fatal(err)
		return
	}

	w.clientsMux.Lock()
	w.clients[conn] = struct{}{}
	w.clientsMux.Unlock()

	w.handleConnection(conn)
}

// handleConnection handles a websocket connection by reading
// messages from the connection and logging them to the console.
//
// This method is called internally when a connection is made to the
// websocket server.
//
// # Parameters:
//
// 	- conn (*websocket.Conn): the websocket connection.
//
// # Example:
//
// 	ws.handleConnection(conn)
func (w *WebSocket) handleConnection(conn *websocket.Conn) {
	defer func() {
		w.clientsMux.Lock()
		delete(w.clients, conn)
		w.clientsMux.Unlock()

		conn.Close()
	}()

	for {
		msgType, msg, err := conn.ReadMessage()
		if err != nil {
			log.Println(err)
			w.clientsMux.Lock()
			delete(w.clients, conn)
			w.clientsMux.Unlock()

			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error reading message: %v", err)
			}
			break
		}

		fmt.Println(msgType)
		fmt.Println(string(msg))
	}
}
