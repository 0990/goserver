package server

import (
	"github.com/0990/goserver/rpc"
	"github.com/0990/goserver/service"
	"github.com/golang/protobuf/proto"
)

type Server struct {
	worker service.Worker
	rpc    *rpc.RPC
}

func NewServer(serverID int32) (*Server, error) {
	p := new(Server)
	p.worker = service.NewWorker()
	rpc, err := rpc.NewRPC(serverID, p.worker)
	if err != nil {
		return nil, err
	}
	p.rpc = rpc
	return p, nil
}

//TODO 添加关闭信号
func (p *Server) Run() {
	p.worker.Run()
	p.rpc.Run()
}

func (p *Server) Post(f func()) {
	p.worker.Post(f)
}

func (p *Server) RegisterRequestHandler(msg proto.Message, f func(rpc.RequestServer, proto.Message)) {
	p.rpc.RegisterRequest(msg, f)
}

func (p *Server) GetServerById(serverID int32) rpc.Server {
	return p.rpc.GetServerById(serverID)
}

func (p *Server) RegisterServerMsg(msg proto.Message, f func(rpc.Server, proto.Message)) {
	p.rpc.RegisterServerMsg(msg, f)
}

func (p *Server) RegisterSessionMsg(msg proto.Message, f func(rpc.Session, proto.Message)) {
	p.rpc.RegisterSessionMsg(msg, f)
}
