package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"strings"
	"sync"
	"time"
)

type message struct {
	sender, text string
	time         time.Time
}

func (m *message) String() string {
	return fmt.Sprintf("%s %s: %s\n", m.time.Format(time.RFC3339), m.sender, m.text)
}

type messageReceiver interface {
	io.WriteCloser
}

type messageBroker struct {
	input chan interface{}
}

func newMessageBroker() *messageBroker {
	mb := new(messageBroker)
	mb.input = make(chan interface{})
	return mb
}

func (m *messageBroker) eventLoop(wg *sync.WaitGroup) {
	defer wg.Done()

	var receivers []messageReceiver

	/*
		serializing messages and joins allows for a single view of messages shared by all clients (sequential consistency).
		it's not necessary and it allows for contention on the main channel but it is simple and expedient.

		for example, if there were two channels for joining and messaging you could join and send a message near enough in
		time that you may or may not see your own message as the runtime will randomly choose which channel to proceed on
		given two blocked channels.

		this way if you block on sending a join and then on sending a message everyone is sure to see the same order of
		events as you.
	*/
	for cmd := range m.input {
		switch v := cmd.(type) {
		case *message:
			if receivers != nil {
				temp := make([]messageReceiver, 0, len(receivers))
				for _, r := range receivers {
					_, err := io.WriteString(r, v.String())
					if err != nil {
						log.Printf("error writing message (%+v): %v", v, err)

						err = r.Close()
						if err != nil {
							log.Printf("error closing message receiver: %v", err)
						}
					} else {
						temp = append(temp, r)
					}
					receivers = temp
				}
			}
		case messageReceiver:
			if v != nil {
				receivers = append(receivers, v)
			}
		}
	}

	for _, r := range receivers {
		err := r.Close()
		if err != nil {
			log.Printf("error closing message receiver: %v", err)
		}
	}
}

func (m *messageBroker) acceptMessage(sender, text string) {
	m.input <- &message{sender, text, time.Now()}
}

func (m *messageBroker) addReceiver(receiver messageReceiver) {
	m.input <- receiver
}

func (m *messageBroker) close() {
	close(m.input)
}

func parseMessage(msg string) (*message, error) {
	i := strings.Index(msg, " ")
	if i == -1 {
		return nil, errors.New("could not find time")
	}

	timeStr, msg := msg[:i], msg[i+1:]
	t, err := time.Parse(time.RFC3339, timeStr)
	if err != nil {
		return nil, fmt.Errorf("could not parse time: %v", err)
	}

	i = strings.Index(msg, ": ")
	if i == -1 {
		return nil, errors.New("could not find sender")
	}

	sender, text := msg[:i], msg[i+2:]

	m := new(message)
	m.time = t
	m.sender = sender
	m.text = strings.TrimSpace(text)

	return m, nil
}
