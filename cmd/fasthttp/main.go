package main

import (
	"github.com/dgrr/fastws"
	routing "github.com/qiangxue/fasthttp-routing"
	"github.com/valyala/fasthttp"
	"log"
	"time"
)

var payload = make([]byte, 12500)

func main() {
	router := routing.New()
	router.Get("/stream", func(ctx *routing.Context) error {
		fastws.Upgrade(func(conn *fastws.Conn) {
			log.Println("client connected")
			defer log.Println("client disconnected")
			defer conn.Close()
			ticker := time.NewTicker(time.Millisecond * 40)
			defer ticker.Stop()
			for {
				<-ticker.C

				_, err := conn.WriteMessage(fastws.ModeBinary, payload)
				if err != nil {
					return
				}
			}
		})(ctx.RequestCtx)
		return nil
	})

	srv := &fasthttp.Server{
		Handler:         router.HandleRequest,
		WriteBufferSize: 512,
		ReadBufferSize:  512,
		IdleTimeout:     10 * time.Second,
		//DisableKeepalive: true,
	}
	//err := srv.ListenAndServe(":4242")
	log.Println("Listening on :4242")
	err := srv.ListenAndServe("0.0.0.0:4242")
	if err != nil {
		log.Fatal(err)
	}
}
