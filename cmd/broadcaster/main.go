package main

import (
	"fmt"
	"github.com/buaazp/fasthttprouter"
	"github.com/dgrr/fastws"
	"github.com/valyala/fasthttp"
	"os"
	"os/signal"
	"sync"
	"time"
)

var payload = make([]byte, 12500)

type Broadcaster struct {
	lck sync.Mutex
	cs  []*fastws.Conn
}

func (b *Broadcaster) Add(c *fastws.Conn) {
	b.lck.Lock()
	b.cs = append(b.cs, c)
	fmt.Println(len(b.cs))
	b.lck.Unlock()
}

func (b *Broadcaster) Start() {
	ticker := time.NewTicker(time.Millisecond * 40)
	defer ticker.Stop()
	for {
		b.lck.Lock()
		now := time.Now()
		for i := 0; i < len(b.cs); i++ {
			c := b.cs[i]
			_, err := c.WriteMessage(fastws.ModeBinary, payload)
			if err != nil {
				b.cs = append(b.cs[:i], b.cs[i+1:]...)
				fmt.Println(err)
				fmt.Println(len(b.cs))
				continue
			}
		}
		fmt.Println("elapsed time:", time.Since(now))
		b.lck.Unlock()

		<-ticker.C
	}
}

func main() {
	b := &Broadcaster{}
	router := fasthttprouter.New()
	router.GET("/stream", fastws.Upgrade(func(c *fastws.Conn) {
		b.Add(c)
		c.ReadTimeout = 0
		for {
			_, _, err := c.ReadMessage(nil)
			if err != nil {
				return
			}
		}
	}))
	go b.Start()

	server := fasthttp.Server{
		Handler: router.Handler,
	}
	go server.ListenAndServe(":4242")

	fmt.Println("Visit http://localhost:4242")

	sigCh := make(chan os.Signal)
	signal.Notify(sigCh, os.Interrupt)
	<-sigCh
	signal.Stop(sigCh)
	signal.Reset(os.Interrupt)
	server.Shutdown()
}
