package main

import (
	"crypto/tls"
	"encoding/hex"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"time"

	"github.com/chzyer/readline"
	"github.com/fatih/color"
	"github.com/gorilla/websocket"
)

type session struct {
	ws      *websocket.Conn
	rl      *readline.Instance
	errChan chan error
}

func connect(url, origin string, rlConf *readline.Config, allowInsecure bool, readOnly bool) error {
	headers := make(http.Header)
	headers.Add("Origin", origin)

	dialer := websocket.Dialer{
		Proxy: http.ProxyFromEnvironment,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: allowInsecure,
		},
	}
	ws, _, err := dialer.Dial(url, headers)
	if err != nil {
		return err
	}

	rl, err := readline.NewEx(rlConf)
	if err != nil {
		return err
	}
	defer rl.Close()

	sess := &session{
		ws:      ws,
		rl:      rl,
		errChan: make(chan error),
	}

	if !readOnly {
		go sess.readConsole()
	}
	go sess.readWebsocket()

	return <-sess.errChan
}

func (s *session) readConsole() {
	for {
		line, err := s.rl.Readline()
		if err != nil {
			fmt.Fprintln(os.Stderr, "s.rl.Readline failed")
			s.errChan <- err
			return
		}

		err = s.ws.WriteMessage(websocket.TextMessage, []byte(line))
		if err != nil {
			fmt.Fprintln(os.Stderr, "s.ws.WriteMessge failed")
			s.errChan <- err
			return
		}
	}
}

func bytesToFormattedHex(bytes []byte) string {
	text := hex.EncodeToString(bytes)
	return regexp.MustCompile("(..)").ReplaceAllString(text, "$1 ")
}

func (s *session) readWebsocket() {
	rxSprintf := color.New(color.FgGreen).SprintfFunc()

	for {
		msgType, buf, err := s.ws.ReadMessage()
		if err != nil {
			fmt.Fprintln(os.Stderr, "s.ws.ReadMessage failed")
			s.errChan <- err
			return
		}

		var text string
		switch msgType {
		case websocket.TextMessage:
			text = string(buf)
		case websocket.BinaryMessage:
			text = bytesToFormattedHex(buf)
		default:
			fmt.Fprintf(os.Stderr, "unknown websocket frame type: %d\n", msgType)
			continue
		}

		fmt.Fprint(s.rl.Stdout(), rxSprintf("%s [%s] %s\n", time.Now().UTC(), s.ws.LocalAddr(), text))
	}
}
