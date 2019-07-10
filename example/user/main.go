package main

import (
	"fmt"
	cmsg "github.com/0990/goserver/example/msg"
	"github.com/0990/goserver/rpc"
	"github.com/0990/goserver/server"
	"github.com/golang/protobuf/proto"
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
	s.RegisterSessionMsg(&cmsg.ReqHello{}, func(client rpc.Session, msg1 proto.Message) {
		req := msg1.(*cmsg.ReqHello)
		fmt.Println(req.Name)
		resp := &cmsg.RespHello{
			Name: "userbaobao",
		}
		client.SendMsg(resp)
	})

	s.RegisterServerMsg(&cmsg.ReqServer2Server{}, func(server rpc.Server, message proto.Message) {
		req := message.(*cmsg.ReqServer2Server)
		fmt.Println("server2server message", req)
	})

	s.RegisterRequestHandler(&cmsg.ReqRequest{}, func(server rpc.RequestServer, message proto.Message) {
		now := time.Now()
		req := message.(*cmsg.ReqRequest)
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
