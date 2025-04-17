package main

import (
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

var payload = make([]byte, 12500)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func streamHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("upgrade failed:", err)
		return
	}
	defer func() {
		conn.Close()
		log.Println("client disconnected")
	}()

	log.Println("client connected")

	ticker := time.NewTicker(40 * time.Millisecond)
	defer ticker.Stop()

	for {
		<-ticker.C
		if err := conn.WriteMessage(websocket.BinaryMessage, payload); err != nil {
			log.Println("write failed:", err)
			return
		}
	}
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/stream", streamHandler)

	srv := &http.Server{
		Addr:    ":4242",
		Handler: r,
	}

	log.Println("Listening on :4242")
	if err := srv.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
