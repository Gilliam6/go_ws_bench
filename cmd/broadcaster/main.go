package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"time"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/mailru/easygo/netpoll"
)

var (
	// global listener and netpoller
	listener net.Listener
	poller   netpoll.Poller

	// connection registry
	connsLock  sync.RWMutex
	conns      = make(map[int]net.Conn)
	nextConnID int

	// worker semaphore for goroutine pool
	workerSem chan struct{}

	// shared payload buffer
	payload = make([]byte, 12500)
)

func main() {
	var err error
	// 1) Create epoll-based poller ([godoc.org](https://godoc.org/github.com/mailru/easygo/netpoll))
	poller, err = netpoll.New(nil)
	if err != nil {
		panic(err)
	}

	// 2) Listen on TCP and zero-copy upgrade for WebSocket ([medium.com](https://medium.com/free-code-camp/million-websockets-and-go-cc58418460bb))
	listener, err = net.Listen("tcp", ":4242")
	if err != nil {
		panic(err)
	}

	// 3) Register listener for read events (edge-triggered)
	descListener := netpoll.Must(netpoll.HandleListener(listener, netpoll.EventRead|netpoll.EventEdgeTriggered))
	poller.Start(descListener, acceptHandler)

	// 4) Initialize worker pool size = NumCPU*2 ([medium.com](https://medium.com/free-code-camp/million-websockets-and-go-cc58418460bb))
	workerSem = make(chan struct{}, runtime.NumCPU()*20)

	// 5) Start broadcast ticker
	ticker := time.NewTicker(40 * time.Millisecond)
	go func() {
		for range ticker.C {
			broadcast(payload)
		}
	}()

	fmt.Println("Listening on :4242")
	// Graceful shutdown
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	<-sig
	listener.Close()
	//poller.()
}

// acceptHandler handles new inbound connections
func acceptHandler(ev netpoll.Event) {
	for {
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		// Zero-copy WebSocket upgrade using gobwas/ws ([medium.com](https://medium.com/free-code-camp/million-websockets-and-go-cc58418460bb))
		_, err = ws.Upgrade(conn)
		if err != nil {
			conn.Close()
			continue
		}
		log.Println("New connection from", conn.RemoteAddr())
		// register and assign unique ID
		id := registerConn(conn)
		// one-shot read registration to avoid persistent goroutine/buffer per conn ([medium.com](https://medium.com/free-code-camp/million-websockets-and-go-cc58418460bb))
		desc := netpoll.Must(netpoll.HandleReadOnce(conn))
		poller.Start(desc, func(ev netpoll.Event) {
			// limit concurrent handlers
			workerSem <- struct{}{}
			go func() {
				defer func() { <-workerSem }()
				if ev&netpoll.EventReadHup != 0 {
					// cleanup on close
					unregisterConn(id)
					poller.Stop(desc)
					conn.Close()
					return
				}
				// handle ping/pong and ignore payload
				//wsutil.ReadServerFrame(conn, nil)
				// resume for next event
				poller.Resume(desc)
			}()
		})
	}
}

// broadcast sends data to all connections using worker pool for writes
func broadcast(data []byte) {
	connsLock.RLock()
	for id, conn := range conns {
		workerSem <- struct{}{}
		go func(id int, c net.Conn) {
			defer func() { <-workerSem }()
			// efficient binary write without extra buffering ([medium.com](https://medium.com/free-code-camp/million-websockets-and-go-cc58418460bb))
			if err := wsutil.WriteServerBinary(c, data); err != nil {
				unregisterConn(id)
				c.Close()
			}
		}(id, conn)
	}
	connsLock.RUnlock()
}

// registerConn safely stores a new connection and returns its ID
type Conn interface{ net.Conn }

func registerConn(c net.Conn) int {
	connsLock.Lock()
	defer connsLock.Unlock()
	id := nextConnID
	nextConnID++
	conns[id] = c
	return id
}

// unregisterConn removes a connection by ID
func unregisterConn(id int) {
	log.Println("Unregistering connection", id)
	connsLock.Lock()
	defer connsLock.Unlock()
	delete(conns, id)
}
