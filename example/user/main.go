package main

import (
	"fmt"
	"github.com/0990/goserver"
	"github.com/0990/goserver/example/msg"
	"github.com/0990/goserver/rpc"
	"github.com/0990/goserver/server"
	"time"
)

func main() {

	s, err := server.NewServer(101, goserver.Config{Nats: "127.0.0.1:4222"})
	if err != nil {
		fmt.Println(err)
		return
	}

	//注册客户端消息事件handler
	s.RegisterSessionMsgHandler(func(client rpc.Session, req *pb.ReqHello) {
		resp := &pb.RespHello{Name: "回复您的请求"}
		fmt.Println("收到客户端来的消息:", req.Name)
		client.SendMsg(resp)
	})

	//注册send事件handler
	s.RegisterServerHandler(func(server rpc.Server, req *pb.ReqSend) {
		fmt.Println("收到gate来的消息:", req.Name)
	})

	//注册request事件handler
	s.RegisterRequestMsgHandler(func(server rpc.RequestServer, req *pb.ReqRequest) {
		resp := &pb.RespRequest{Name: "我是request返回消息"}
		fmt.Println("收到gate来的request消息:", req.Name)
		server.Answer(resp)
	})

	//注册call事件handler
	s.RegisterRequestMsgHandler(func(server rpc.RequestServer, req *pb.ReqCall) {
		resp := &pb.RespCall{Name: "我是call返回消息"}
		fmt.Println("收到gate来的call消息:", req.Name)
		server.Answer(resp)
	})

	s.Run()
	time.Sleep(time.Hour)
}
