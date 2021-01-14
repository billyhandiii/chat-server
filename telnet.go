package main

import (
	"bufio"
	"context"
	"io"
	"log"
	"net"
	"strings"
	"sync"
	"time"
)

func acceptConnections(ctx context.Context, wg *sync.WaitGroup, cfg *config, mb *messageBroker) {
	defer wg.Done()

	laddr := new(net.TCPAddr)

	if len(cfg.IP) > 0 {
		ip := net.ParseIP(cfg.IP)
		if ip == nil {
			log.Printf("ip address could not be parsed (%v); binding to all interfaces", cfg.IP)
		}

		if cfg.Port < 0 {
			log.Printf("port cannot be negative; picking a random one")
		}

		laddr.IP = ip
		laddr.Port = cfg.Port
	}

	tl, err := net.ListenTCP("tcp", laddr)
	if err != nil {
		log.Fatalf("could not start listener: %v", err)
	}
	go func() {
		<-ctx.Done()
		tl.Close()
	}()

	log.Printf("listening for messages over telnet on %v", tl.Addr())

	initialWait := time.Millisecond
	maxWait := time.Second
	waitOnError := time.Millisecond
	for {
		select {
		case <-ctx.Done():
			return
		default:
			conn, err := tl.AcceptTCP()
			if err != nil {
				/*
					some of these errors are likely recoverable and some not.  rather than differentiate at this time we are
					just going to exponentially back off and assume the user eventually kills the server.  that way we don't
					peg the cpu in an fatal situation but we also don't bail out of recoverable situations.
				*/
				if waitOnError > initialWait {
					log.Print("error accepting connection; waiting ", waitOnError)
				}
				waitOnError *= 2
				if waitOnError > maxWait {
					waitOnError = maxWait
				}
				time.Sleep(waitOnError)
				continue
			}
			waitOnError = initialWait // reset error wait

			go readMessagesFromConnection(conn, mb)
		}
	}
}

func readMessagesFromConnection(c *net.TCPConn, mb *messageBroker) {
	scanner := bufio.NewScanner(c)
	defer c.Close()

	const (
		enterNamePrompt        = "Enter your name and press enter:\n"
		nameCannotContainColon = "Your name cannot contain a colon (:) character!\n"
	)

	var (
		sender string
		msg    string
	)

	_, err := io.WriteString(c, enterNamePrompt)
	if err != nil {
		return
	}

	for scanner.Scan() {
		msg = strings.TrimSpace(scanner.Text())

		if len(msg) > 0 {
			if len(sender) == 0 {
				if strings.Index(msg, ":") == -1 {
					sender = msg
					mb.addReceiver(c)
				} else {
					_, err := io.WriteString(c, nameCannotContainColon)
					if err != nil {
						return
					}
				}
			} else {
				mb.acceptMessage(sender, msg)
			}
		}

		if len(sender) == 0 {
			_, err := io.WriteString(c, enterNamePrompt)
			if err != nil {
				return
			}
		}
	}
}
