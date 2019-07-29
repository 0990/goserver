package main

import (
	"fmt"
	cmsg "github.com/0990/goserver/example/msg"
	"github.com/0990/goserver/rpc"
	"github.com/0990/goserver/server"
	"time"
)

func main() {

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

	s.RegisterServerHandler(func(server rpc.Server, req *cmsg.ReqServer2Server) {
		fmt.Println("sendtoserver data:", req)
	})

	s.RegisterRequestMsgHandler(func(server rpc.RequestServer, req *cmsg.ReqRequest) {
		fmt.Println("request req data:", req)
		resp := &cmsg.RespRequest{
			Name: req.Name,
		}
		server.Answer(resp)
	})

	s.Run()
	time.Sleep(time.Hour)
}
