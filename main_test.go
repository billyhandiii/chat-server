package main

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/stretchr/testify/require"
)

const (
	unitTestConfigPath = "test_data/configs/unit_test_config.json"
	chatPort           = 30405
	httpPort           = 30406
)

func TestTwoClientsSeeEachOthersMessages(t *testing.T) {
	const (
		sender1 = "wario"
		text1   = "wrahahaha!"
		sender2 = "waluigi"
		text2   = "waaaahh!"
	)

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	go startChatServer(ctx, []string{"", unitTestConfigPath}, discardCloser{})

	// connect senders
	sender1Conn, err := net.Dial("tcp", fmt.Sprintf("127.0.0.1:%v", chatPort))
	require.NoError(t, err)
	sender2Conn, err := net.Dial("tcp", fmt.Sprintf("127.0.0.1:%v", chatPort))
	require.NoError(t, err)

	// read messages for senders
	sender1Msgs := readMessagesForTest(t, sender1Conn)
	sender2Msgs := readMessagesForTest(t, sender2Conn)

	// send sender's names
	_, err = io.WriteString(sender1Conn, sender1+"\n")
	require.NoError(t, err)
	_, err = io.WriteString(sender2Conn, sender2+"\n")
	require.NoError(t, err)

	time.Sleep(time.Millisecond)

	// send sender's messages
	_, err = io.WriteString(sender1Conn, text1+"\n")
	require.NoError(t, err)
	_, err = io.WriteString(sender2Conn, text2+"\n")
	require.NoError(t, err)

	var sender1Count, text1Count, sender2Count, text2Count int

	for _, m := range []*message{<-sender1Msgs, <-sender2Msgs, <-sender1Msgs, <-sender2Msgs} {
		if m.sender == sender1 {
			sender1Count++
		} else if m.sender == sender2 {
			sender2Count++
		}

		if m.text == text1 {
			text1Count++
		} else if m.text == text2 {
			text2Count++
		}
	}

	assert.Equal(t, sender1Count, 2)
	assert.Equal(t, text1Count, 2)
	assert.Equal(t, sender2Count, 2)
	assert.Equal(t, text2Count, 2)
}

func TestClientCanConnectAndSendAMessage(t *testing.T) {
	const (
		sender = "billy"
		text   = "wooooowwweeee!"
	)

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	go startChatServer(ctx, []string{"", unitTestConfigPath}, discardCloser{})

	c, err := net.Dial("tcp", fmt.Sprintf("127.0.0.1:%v", chatPort))
	require.NoError(t, err)

	msgs := readMessagesForTest(t, c)

	_, err = io.WriteString(c, sender+"\n")
	require.NoError(t, err)

	_, err = io.WriteString(c, text+"\n")
	require.NoError(t, err)

	m := <-msgs
	require.NotNil(t, m)
	require.Equal(t, m.sender, sender)
	require.Equal(t, m.text, text)
}

func TestMessagesAreLoggedToAFile(t *testing.T) {
	const (
		sender = "billy"
		text   = "yeeee hawwww!"
	)

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	logRdr, logWtr := io.Pipe()

	go startChatServer(ctx, []string{"", unitTestConfigPath}, logWtr)

	msgs := readMessagesForTest(t, logRdr)

	_, err := http.Post(
		fmt.Sprintf("http://localhost:%v/", httpPort),
		"application/json",
		strings.NewReader(fmt.Sprintf(`{"sender": "%s", "text": "%s"}`, sender, text)),
	)
	require.NoError(t, err)

	m := <-msgs
	require.NotNil(t, m)
	require.Equal(t, m.sender, sender)
	require.Equal(t, m.text, text)
}

type discardCloser struct{}

func (d discardCloser) Write(bs []byte) (int, error) {
	return len(bs), nil
}

func (d discardCloser) Close() error {
	return nil
}
