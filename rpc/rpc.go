package rpc

import (
	"github.com/0990/goserver/service"
	"github.com/golang/protobuf/proto"
	"sync"
)

type RPC struct {
	sync.Mutex
	sid2server map[int32]Server
	rpcClient  *RPCClient
	serverid   int32
	worker     service.Worker
}

func NewRPC(serverid int32, worker service.Worker) (*RPC, error) {
	rpcClient, err := newRPCClient(serverid)
	if err != nil {
		return nil, err
	}
	p := new(RPC)
	p.worker = worker
	p.serverid = serverid
	p.rpcClient = rpcClient
	return p, nil
}

func (p *RPC) Run() {
	p.rpcClient.Run()
	//p.worker.Run()
}

func (p *RPC) RegisterMsg(msg proto.Message, f MsgHandler) {
	p.rpcClient.processor.RegisterMsg(msg, f)
}

func (p *RPC) RegisterRequest(msg proto.Message, f RequestHandler) {
	p.rpcClient.processor.RegisterRequest(msg, f)
}

func (p *RPC) GetServerById(serverid int32) Server {
	p.Lock()
	defer p.Unlock()
	if v, ok := p.sid2server[serverid]; ok {
		return v
	}
	s := NewServer(p.rpcClient, serverid)
	p.sid2server[serverid] = s
	return s
}

//TODO add
func (p *RPC) GetServerByType(serverType ServerType) Server {
	return nil
}

func (p *RPC) RegisterSend2Session(send2Session func(sesid int32, msgid uint16, data []byte)) {
	p.rpcClient.RegisterSend2Session(send2Session)
}
