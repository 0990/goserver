package server

import (
	"github.com/0990/goserver/network"
	"github.com/0990/goserver/rpc"
	"github.com/0990/goserver/service"
	"github.com/golang/protobuf/proto"
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

func (p *Gate) RegisterNetWorkEvent(onNew, onClose func(conn network.Session)) {
	p.networkMgr.RegisterEvent(onNew, onClose)
}

func (p *Gate) RegisterSessionHandler(msg proto.Message, f func(network.Session, proto.Message)) {
	p.networkMgr.RegisterSessionHandler(msg, f)
}

func (p *Gate) RegisterRequestHandler(msg proto.Message, f func(rpc.RequestServer, proto.Message)) {
	p.rpc.RegisterRequest(msg, f)
}

func (p *Gate) GetServerById(serverID int32) rpc.Server {
	return p.rpc.GetServerById(serverID)
}

func (p *Gate) GetSession(sesID int32) (network.Session, bool) {
	return p.networkMgr.GetSession(sesID)
}
