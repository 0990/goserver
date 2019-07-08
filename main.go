package main

import (
	"flag"
	"fmt"
	"github.com/0990/goserver/gate"
	cmsg "github.com/0990/goserver/msg"
	"github.com/0990/goserver/network"
	"github.com/golang/protobuf/proto"
	"time"
)

var addr = flag.String("addr", "localhost:8080", "http service address")

type clientMgr struct {
	clients map[network.Session]struct{}
}

func main() {
	mgr := &clientMgr{
		clients: make(map[network.Session]struct{}),
	}

	g, err := gate.NewGate(*addr)
	if err != nil {
		fmt.Println(err)
		return
	}

	g.RegisterNetWorkEvent(func(conn network.Session) {
		mgr.clients[conn] = struct{}{}
	}, func(conn network.Session) {
		delete(mgr.clients, conn)
	})
	g.RegisterSessionHandler(&cmsg.ReqHello{}, func(client network.Session, msg1 proto.Message) {
		req := msg1.(*cmsg.ReqHello)
		fmt.Println(req.Name)
		resp := &cmsg.RespHello{
			Name: "baobao",
		}
		client.SendMsg(resp)
	})
	g.Run()

	ticker := time.NewTicker(time.Second)
	for range ticker.C {
		g.Post(func() {
			fmt.Println(mgr.clients)
		})
	}
}
