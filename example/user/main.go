package main

import (
	"fmt"
	"github.com/0990/goserver/example/msg"
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

	//注册客户端消息事件handler
	s.RegisterSessionMsgHandler(func(client rpc.Session, req *pb.ReqHello) {
		resp := &pb.RespHello{Name: "回复您的请求"}
		client.SendMsg(resp)
	})

	//注册send事件handler
	s.RegisterServerHandler(func(server rpc.Server, req *pb.ReqSend) {
		fmt.Println("收到消息:", req.Name)
	})

	//注册request事件handler
	s.RegisterRequestMsgHandler(func(server rpc.RequestServer, req *pb.ReqRequest) {
		resp := &pb.RespRequest{Name: "我是request返回消息"}
		server.Answer(resp)
	})

	//注册call事件handler
	s.RegisterRequestMsgHandler(func(server rpc.RequestServer, req *pb.ReqCall) {
		resp := &pb.RespCall{Name: "我是call返回消息"}
		server.Answer(resp)
	})

	s.Run()
	time.Sleep(time.Hour)
}
