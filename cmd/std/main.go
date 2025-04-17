package main

import (
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	_ "net/http/pprof"
)

var (
	payload  = make([]byte, 12500)
	upgrader = websocket.Upgrader{
		ReadBufferSize:  512,
		WriteBufferSize: 512,
		CheckOrigin:     func(r *http.Request) bool { return true },
	}
	pingPeriod = 30 * time.Second
)

func streamHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("upgrade:", err)
		return
	}
	defer func() {
		conn.Close()
		log.Println("client disconnected")
	}()
	log.Println("client connected")

	// ———————————————
	// настройка дедлайнов и ping/pong
	conn.SetReadLimit(512)                           // макс. размер сообщения
	conn.SetReadDeadline(time.Now().Add(pingPeriod)) // initial
	conn.SetPongHandler(func(string) error {
		// продлеваем дедлайн при получении pong
		return conn.SetReadDeadline(time.Now().Add(pingPeriod))
	})

	// запустим «читателя», чтобы не захлебнуться контролами и возможными ошибками
	go func() {
		for {
			if _, _, err := conn.NextReader(); err != nil {
				// когда клиент закрылся — выходим
				return
			}
			// здесь можно считать текстовые/бинарные сообщения,
			// но если они не нужны — просто NextReader() «съест» их
		}
	}()

	// ———————————————
	// пишем в цикле и шлём ping, чтобы поддерживать alive
	ticker := time.NewTicker(40 * time.Millisecond)
	pingTicker := time.NewTicker(pingPeriod)
	defer ticker.Stop()
	defer pingTicker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := conn.WriteMessage(websocket.BinaryMessage, payload); err != nil {
				return
			}
		case <-pingTicker.C:
			// отправляем ping, чтобы клиент ответил pong
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func main() {
	http.HandleFunc("/stream", streamHandler)
	srv := &http.Server{
		Addr:        ":4242",
		Handler:     http.DefaultServeMux,
		IdleTimeout: 10 * time.Second,
		//DisableKeepAlives: true,
	}
	log.Println("Listening on :4242")
	log.Fatal(srv.ListenAndServe())
	//log.Fatal(http.ListenAndServe(":4242", nil))
}
