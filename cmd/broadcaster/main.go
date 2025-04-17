package main

import (
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"time"

	"github.com/buaazp/fasthttprouter"
	"github.com/dgrr/fastws"
	"github.com/valyala/fasthttp"
)

var payload = make([]byte, 12500)

type Pool struct {
	id  int
	lck sync.Mutex
	cs  []*fastws.Conn
}

func (p *Pool) Start() {
	ticker := time.NewTicker(40 * time.Millisecond)
	defer ticker.Stop()
	var now time.Time
	for range ticker.C {
		p.lck.Lock()
		now = time.Now()
		for i := 0; i < len(p.cs); {
			c := p.cs[i]
			if _, err := c.WriteMessage(fastws.ModeBinary, payload); err != nil {
				p.cs = append(p.cs[:i], p.cs[i+1:]...)
				fmt.Printf("Pool %d: removed conn, left %d\n", p.id, len(p.cs))
			} else {
				i++
			}
		}
		fmt.Printf("Pool %d: elapsed time %v\n", p.id, time.Since(now))
		p.lck.Unlock()
	}
}

type Broadcaster struct {
	poolLock sync.Mutex
	pools    []*Pool
	poolSize int
}

func (b *Broadcaster) Add(c *fastws.Conn) {
	b.poolLock.Lock()
	defer b.poolLock.Unlock()

	for idx, pool := range b.pools {
		pool.lck.Lock()
		if len(pool.cs) < b.poolSize {
			pool.cs = append(pool.cs, c)
			pool.lck.Unlock()
			fmt.Printf("Added to pool %d, size %d\n", idx, len(pool.cs))
			return
		}
		pool.lck.Unlock()
	}

	pid := len(b.pools)
	newPool := &Pool{id: pid, cs: make([]*fastws.Conn, 0, 50)}
	newPool.cs = append(newPool.cs, c)
	b.pools = append(b.pools, newPool)
	fmt.Printf("Created pool %d, size %d\n", pid, len(newPool.cs))

	go newPool.Start()
}

func main() {
	size, err := strconv.Atoi(os.Args[0])
	if err != nil {
		fmt.Printf("args should be int: %s\n", err)
		os.Exit(1)
	}
	b := &Broadcaster{poolSize: size}

	router := fasthttprouter.New()
	router.GET("/stream", fastws.Upgrade(func(c *fastws.Conn) {
		c.ReadTimeout = 0
		b.Add(c)
		for {
			if _, _, err := c.ReadMessage(nil); err != nil {
				return
			}
		}
	}))

	server := fasthttp.Server{Handler: router.Handler}

	go func() {
		if err := server.ListenAndServe(":4242"); err != nil {
			fmt.Println("Server error:", err)
		}
	}()

	fmt.Println("Visit http://localhost:4242")

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)
	<-sigCh

	server.Shutdown()
}
