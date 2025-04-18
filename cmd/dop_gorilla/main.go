package main

import (
	"log"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
)

const (
	addr         = ":4242"
	maxClients   = 10000 // макс. ожидаемое число клиентов
	sendInterval = 40 * time.Millisecond
	pingInterval = 30 * time.Second
	bufferSize   = 512 // минимальный размер буфера
)

var (
	// заранее подготавливаем сообщение
	payload = make([]byte, 12500)
	//prepared *websocket.PreparedMessage

	upgrader = websocket.Upgrader{
		ReadBufferSize:  bufferSize,
		WriteBufferSize: bufferSize,
		CheckOrigin:     func(r *http.Request) bool { return true },
	}

	// срез для коннектов, atomic-access для индексации
	clients    = make([]*websocket.Conn, maxClients)
	clientCnt  int32 // реальный счётчик занятых слотов
	register   = make(chan *websocket.Conn, 100)
	unregister = make(chan int, 100)
)

//func init() {
//var err error
//prepared, err = websocket.NewPreparedMessage(websocket.BinaryMessage, payload)
//if err != nil {
//	log.Fatalf("failed to prepare message: %v", err)
//}
//}

func streamHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("upgrade:", err)
		return
	}
	// REGISTER: отправляем указатель на центральный диспетчер
	register <- conn
}

func dispatcher() {
	ticker := time.NewTicker(sendInterval)
	pingT := time.NewTicker(pingInterval)
	defer ticker.Stop()
	defer pingT.Stop()

	for {
		select {
		case conn := <-register:
			// найти первый пустой слот
			idx := int(atomic.AddInt32(&clientCnt, 1) - 1)
			if idx < maxClients {
				clients[idx] = conn
				// настраиваем пинг/понг для этого коннекта
				conn.SetReadDeadline(time.Now().Add(pingInterval))
				conn.SetPongHandler(func(string) error {
					return conn.SetReadDeadline(time.Now().Add(pingInterval))
				})
				log.Println(idx, " client connected")
			} else {
				// переполнен пул — сброс
				atomic.AddInt32(&clientCnt, -1)
				conn.Close()
			}

		case idx := <-unregister:
			if idx >= 0 && idx < int(atomic.LoadInt32(&clientCnt)) {
				// закрываем и освобождаем слот
				clients[idx].Close()
				// перекидываем последний клиент на место удалённого
				last := int(atomic.AddInt32(&clientCnt, -1))
				clients[idx] = clients[last]
				clients[last] = nil
				log.Println(last, " client disconnected")
			}

		case <-ticker.C:
			// отправляем всем
			limit := int(atomic.LoadInt32(&clientCnt))
			for i := 0; i < limit; i++ {
				if err := clients[i].WriteMessage(websocket.BinaryMessage, payload); err != nil {
					// при ошибке — регистрируем удаление
					unregister <- i
				}
			}

		case <-pingT.C:
			// шлём пинг всем
			limit := int(atomic.LoadInt32(&clientCnt))
			for i := 0; i < limit; i++ {
				_ = clients[i].WriteMessage(websocket.PingMessage, nil)
			}
		}
	}
}

func main() {
	http.HandleFunc("/stream", streamHandler)

	go dispatcher()

	log.Printf("Listening on %s\n", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal(err)
	}
}
