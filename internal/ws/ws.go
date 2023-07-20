package ws

import (
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

type WebSocket struct {
	clients    map[*websocket.Conn]struct{}
	clientsMux sync.Mutex
}

func NewWebSocket() *WebSocket {
	return &WebSocket{
		clients: make(map[*websocket.Conn]struct{}),
	}
}

func (w *WebSocket) Start() {
	http.HandleFunc("/listen", w.websocketHandler)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal(err)
	}
}

func (w *WebSocket) Stop() {
	w.clientsMux.Lock()
	defer w.clientsMux.Unlock()

	for client := range w.clients {
		client.Close()
	}

	w.clients = make(map[*websocket.Conn]struct{})
}

func (w *WebSocket) DispatchUpdate(update string) {
	w.clientsMux.Lock()
	defer w.clientsMux.Unlock()

	for client := range w.clients {
		err := client.WriteMessage(websocket.TextMessage, []byte(update))
		if err != nil {
			log.Println(err)
			return
		}
	}
}

func (w *WebSocket) websocketHandler(res http.ResponseWriter, req *http.Request) {
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) (bool) { return true },
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
