package main

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"sync"
)

type httpMessage struct {
	Sender string `json:"sender"`
	Text   string `json:"text"`
}

func acceptHTTPMessages(ctx context.Context, wg *sync.WaitGroup, cfg *config, mb *messageBroker) {
	defer wg.Done()

	if cfg.HTTP != nil {
		httpLAddr := new(net.TCPAddr)

		ip := net.ParseIP(cfg.HTTP.IP)
		if ip == nil {
			log.Printf("http ip address could not be parsed (%v); binding to all interfaces", cfg.HTTP.IP)
		}

		if cfg.HTTP.Port < 0 {
			log.Printf("http port cannot be negative; picking a random one")
		}

		httpLAddr.IP = ip
		httpLAddr.Port = cfg.HTTP.Port

		httpL, err := net.ListenTCP("tcp", httpLAddr)
		if err != nil {
			log.Fatalf("could not start http listener: %v", err)
		}

		go func() {
			<-ctx.Done()
			httpL.Close()
		}()

		log.Printf("listening for messages over http on %v", httpL.Addr())
		err = http.Serve(httpL, http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			status := 400

			defer func() {
				writer.WriteHeader(status)
			}()

			bs, err := ioutil.ReadAll(request.Body)
			if err != nil {
				log.Printf("http body couldn't be read: %v", err)
				return
			}

			var msg httpMessage

			err = json.Unmarshal(bs, &msg)
			if err != nil {
				log.Printf("http body couldn't be parsed: %v", err)
				return
			}

			mb.acceptMessage(msg.Sender, msg.Text)

			status = http.StatusAccepted
		}))

		if err != nil {
			log.Printf("http listener stopped: %v", err)
		}
	}
}
