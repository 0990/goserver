package main

import (
	"fmt"
	cmsg "github.com/0990/goserver/example/msg"
	"github.com/0990/goserver/rpc"
	"github.com/0990/goserver/server"
	"time"
)

type clientMgr struct {
	clients map[rpc.Session]struct{}
}

func main() {
	//mgr := &clientMgr{
	//	clients: make(map[rpc.Session]struct{}),
	//}

	s, err := server.NewServer(101)
	if err != nil {
		fmt.Println(err)
		return
	}

	//测试sessionHandler
	s.RegisterSessionMsgHandler(func(client rpc.Session, req *cmsg.ReqHello) {
		fmt.Println(req.Name)
		resp := &cmsg.RespHello{
			Name: "userbaobao",
		}
		client.SendMsg(resp)
	})

	//s.RegisterServerMsg(&cmsg.ReqServer2Server{}, func(server rpc.Server, message proto.Message) {
	//	req := message.(*cmsg.ReqServer2Server)
	//	fmt.Println("server2server message", req)
	//})

	s.RegisterServerHandler(func(server rpc.Server, req *cmsg.ReqServer2Server) {
		fmt.Println("server2server message", req)
	})

	s.RegisterRequestMsgHandler(func(server rpc.RequestServer, req *cmsg.ReqRequest) {
		now := time.Now()
		fmt.Println("request message", req)
		resp := &cmsg.RespRequest{
			Name: req.Name,
		}
		server.Answer(resp)
		fmt.Println("user reqrequest", time.Since(now))
	})

	s.Run()
	time.Sleep(time.Hour)
}
