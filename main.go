package main

import (
	"flag"
	"fmt"
	"github.com/0990/goserver/gate"
	"time"
)

var addr = flag.String("addr", "localhost:8080", "http service address")

type clientMgr struct {
	clients map[*gate.Client]struct{}
}

func main() {
	mgr := &clientMgr{
		clients: make(map[*gate.Client]struct{}),
	}

	g := gate.NewGate(*addr)
	g.RegisterSessionEvent(func(conn *gate.Client) {
		mgr.clients[conn] = struct{}{}
	}, func(conn *gate.Client) {
		delete(mgr.clients, conn)
	})
	g.Run()

	ticker := time.NewTicker(time.Second)
	for range ticker.C {
		g.Post(func() {
			fmt.Println(mgr.clients)
		})
	}
}
