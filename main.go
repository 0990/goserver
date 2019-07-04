package main

import (
	"flag"
	"fmt"
	"github.com/0990/goserver/gate"
	cmsg "github.com/0990/goserver/msg"
	"github.com/golang/protobuf/proto"
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

	g, err := gate.NewGate(*addr)
	if err != nil {
		fmt.Println(err)
		return
	}
	g.RegisterSessionEvent(func(conn *gate.Client) {
		mgr.clients[conn] = struct{}{}
	}, func(conn *gate.Client) {
		delete(mgr.clients, conn)
	})
	g.RegisterSessionHandler(&cmsg.ReqHello{}, func(client *gate.Client, msg1 proto.Message) {
		req := msg1.(*cmsg.ReqHello)
		fmt.Println(req.Name)
		resp := &cmsg.RespHello{
			Name: "baobao",
		}
		client.SendMsg(resp)
		g.GetServerById(101).Route(client.ReadLoop())
	})
	g.Run()

	ticker := time.NewTicker(time.Second)
	for range ticker.C {
		g.Post(func() {
			fmt.Println(mgr.clients)
		})
	}
}
