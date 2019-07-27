package server

import (
	"github.com/0990/goserver/network"
	"github.com/0990/goserver/rpc"
	"github.com/0990/goserver/service"
	"github.com/golang/protobuf/proto"
	"time"
)

type Gate struct {
	worker     service.Worker
	rpc        *rpc.RPC
	networkMgr *network.Mgr
}

func NewGate(serverID int32, addr string) (*Gate, error) {
	p := new(Gate)
	p.worker = service.NewWorker()
	p.networkMgr = network.NewMgr(addr, p.worker)
	rpc, err := rpc.NewRPC(serverID, p.worker)
	if err != nil {
		return nil, err
	}
	p.rpc = rpc
	p.rpc.RegisterSend2Session(func(sesID int32, msgID uint32, data []byte) {
		if ses, ok := p.GetSession(sesID); ok {
			ses.SendRawMsg(msgID, data)
		}
	})
	return p, nil
}

//TODO 添加关闭信号
func (p *Gate) Run() {
	p.worker.Run()
	p.rpc.Run()
	p.networkMgr.Run()
}

func (p *Gate) Post(f func()) {
	p.worker.Post(f)
}

func (p *Gate) AfterPost(duration time.Duration, f func()) {
	p.worker.AfterPost(duration, f)
}

func (p *Gate) RegisterNetWorkEvent(onNew, onClose func(conn network.Session)) {
	p.networkMgr.RegisterEvent(onNew, onClose)
}

func (p *Gate) RegisterSessionMsgHandler(cb interface{}) {
	p.networkMgr.RegisterSessionMsgHandler(cb)
}

func (p *Gate) RegisterRequestMsgHandler(cb interface{}) {
	p.rpc.RegisterRequestMsgHandler(cb)
}

func (p *Gate) RegisterServerHandler(cb interface{}) {
	p.rpc.RegisterServerMsgHandler(cb)
}

func (p *Gate) GetServerById(serverID int32) rpc.Server {
	return p.rpc.GetServerById(serverID)
}

func (p *Gate) GetSession(sesID int32) (network.Session, bool) {
	return p.networkMgr.GetSession(sesID)
}

func (p *Gate) RouteSessionMsg(msg proto.Message, serverID int32) {
	p.networkMgr.RegisterRawSessionMsgHandler(msg, func(s network.Session, msg proto.Message) {
		p.GetServerById(serverID).RouteSession2Server(s.ID(), msg)
	})
}

func (p *Gate) RegisterRawSessionMsgHandler(msg proto.Message, f func(s network.Session, message proto.Message)) {
	p.networkMgr.RegisterRawSessionMsgHandler(msg, f)
}
