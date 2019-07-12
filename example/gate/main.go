package main

import (
	"flag"
	"fmt"
	cmsg "github.com/0990/goserver/example/msg"
	"github.com/0990/goserver/network"
	"github.com/0990/goserver/server"
	"github.com/0990/goserver/util"
	"github.com/sirupsen/logrus"
	_ "net/http/pprof"
	"time"
)

var addr = flag.String("addr", "localhost:8080", "http service address")

type clientMgr struct {
	clients map[network.Session]struct{}
}

func main() {

	//go func() {
	//	fmt.Println(http.ListenAndServe("0.0.0.0:8888", nil))
	//}()

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
	//g.RegisterSessionMsgHandler(func(session network.Session, msg *cmsg.ReqHello) {
	//	//req := msg1.(*cmsg.ReqHello)
	//	//fmt.Println(req.Name)
	//	//resp := &cmsg.RespHello{
	//	//	Name: "baobao",
	//	//}
	//	//client.SendMsg(resp)
	//	util.PrintGoroutineID("gate receive msg")
	//	//	fmt.Println("gate 收到消息")
	//	g.GetServerById(101).RouteSession2Server(session.ID(), msg)
	//})

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
	fmt.Println(resp)

	//request
	now := time.Now()
	util.PrintCurrNano("request before")
	g.GetServerById(101).Request(&cmsg.ReqRequest{
		Name: "request",
	}, func(resp *cmsg.RespRequest, e error) {
		util.PrintCurrNano("request after")
		fmt.Println("request", time.Since(now))
		if e != nil {
			logrus.Error(e)
			return
		}
		fmt.Println(resp)
		util.PrintGoroutineID("request receive msg")
	})

	ticker := time.NewTicker(time.Millisecond * 100)
	for range ticker.C {
		//now := time.Now()
		//g.GetServerById(101).Request(&cmsg.ReqRequest{
		//	Name: "request",
		//}, func(message proto.Message, e error) {
		//	fmt.Println("request", time.Since(now))
		//	if e != nil {
		//		logrus.Error(e)
		//		return
		//	}
		//	resp := message.(*cmsg.RespRequest)
		//	fmt.Println(resp)
		//})
		//
		//g.Post(func() {
		//	fmt.Println(mgr.clients)
		//})
	}
}
