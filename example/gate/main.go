package main

import (
	"flag"
	"fmt"
	cmsg "github.com/0990/goserver/example/msg"
	"github.com/0990/goserver/network"
	"github.com/0990/goserver/server"
	"github.com/sirupsen/logrus"
	_ "net/http/pprof"
	"time"
)

var addr = flag.String("addr", "0.0.0.0:8080", "http service address")

type clientMgr struct {
	clients map[network.Session]struct{}
}

func main() {

	mgr := &clientMgr{
		clients: make(map[network.Session]struct{}),
	}

	g, err := server.NewGate(100, *addr)
	if err != nil {
		fmt.Println(err)
		return
	}

	g.RegisterNetWorkEvent(func(conn network.Session) {
		mgr.clients[conn] = struct{}{}
	}, func(conn network.Session) {
		delete(mgr.clients, conn)
	})
	g.RouteSessionMsg((*cmsg.ReqHello)(nil), 101)

	g.Run()

	//send
	g.GetServerById(101).Send(&cmsg.ReqServer2Server{
		Name: "server2server xu",
	})

	//call
	resp := &cmsg.RespRequest{}
	err = g.GetServerById(101).Call(&cmsg.ReqRequest{
		Name: "call",
	}, resp)
	if err != nil {
		logrus.WithError(err).Error("error")
		return
	}
	fmt.Println("call return data:", resp)

	//request
	g.GetServerById(101).Request(&cmsg.ReqRequest{
		Name: "request",
	}, func(resp *cmsg.RespRequest, e error) {
		if e != nil {
			logrus.Error(e)
			return
		}
		fmt.Println("request return data", resp)
	})

	time.Sleep(time.Hour)
}
