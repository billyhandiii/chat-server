package main

import (
	"bufio"
	"io"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func readMessagesForTest(t *testing.T, rdr io.ReadCloser) chan *message {
	msgs := make(chan *message)

	go func() {
		scanner := bufio.NewScanner(rdr)

		var msg string
		for scanner.Scan() {
			msg = scanner.Text()
			if len(msg) == 0 {
				time.Sleep(time.Millisecond)
			}

			require.NotEmpty(t, msg)
			if !strings.HasPrefix(msg, "Enter") {
				m, err := parseMessage(msg)
				require.NoError(t, err)
				msgs <- m
			}
		}

		require.NoError(t, scanner.Err())
	}()

	return msgs
}

func TestMessagesAreDisplayedWithTimestampsAndSender(t *testing.T) {
	const (
		sender = "billy"
		text   = "hello, chat server!"
	)
	var wg sync.WaitGroup

	mb := newMessageBroker()

	wg.Add(1)
	go mb.eventLoop(&wg)

	rdr, wtr := io.Pipe()

	mb.addReceiver(wtr)
	mb.acceptMessage(sender, text)
	mb.close()

	msgs := readMessagesForTest(t, rdr)

	m := <-msgs
	require.NotNil(t, m)
	require.Equal(t, m.sender, sender)
	require.Equal(t, m.text, text)

	wg.Wait()
}
