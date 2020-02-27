package main

import (
	"flag"
	cmsg "github.com/0990/goserver/example/msg"
	"github.com/0990/goserver/network"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"net/url"
	"os"
	"os/signal"
	"time"
)

var addr = flag.String("addr", "localhost:8080", "http service address")
var processor = network.NewProcessor()

func main() {
	flag.Parse()

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	u := url.URL{Scheme: "ws", Host: *addr, Path: ""}
	logrus.Infoln("connecting to %s", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		logrus.WithError(err).Fatal("dial")
	}
	defer c.Close()

	done := make(chan struct{})

	go func() {
		defer close(done)
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				logrus.WithError(err).Error("read")
				return
			}
			logrus.WithField("message", message).Debug("recv")
		}
	}()

	ticker := time.NewTicker(time.Second * 2)
	defer ticker.Stop()

	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			req := &cmsg.ReqHello{
				Name: "xujialong",
			}
			data, _ := processor.Marshal(req)
			err := c.WriteMessage(websocket.BinaryMessage, data)
			if err != nil {
				logrus.Println("write:", err)
				return
			}
		case <-interrupt:
			logrus.Println("interrupt")

			// Cleanly close the connection by sending a close message and then
			// waiting (with timeout) for the server to close the connection.
			err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				logrus.Println("write close:", err)
				return
			}
			select {
			case <-done:
			case <-time.After(time.Second):
			}
			return
		}
	}
}
