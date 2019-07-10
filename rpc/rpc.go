package rpc

import (
	"github.com/0990/goserver/service"
	"github.com/golang/protobuf/proto"
	"sync"
)

type RPC struct {
	sync.Mutex
	sid2server map[int32]Server
	client     *Client
	serverID   int32
	worker     service.Worker
}

func NewRPC(serverID int32, worker service.Worker) (*RPC, error) {
	rpcClient, err := newClient(serverID, worker)
	if err != nil {
		return nil, err
	}
	p := new(RPC)
	p.worker = worker
	p.serverID = serverID
	p.client = rpcClient
	p.sid2server = make(map[int32]Server)
	return p, nil
}

//func (p *RPC) Init(serverID int32, worker service.Worker) error {
//	rpcClient, err := newClient(serverID)
//	if err != nil {
//		return err
//	}
//	p.worker = worker
//	p.serverID = serverID
//	p.client = rpcClient
//	return nil
//}

func (p *RPC) Run() {
	p.client.Run()
	//p.worker.Run()
}

func (p *RPC) RegisterServerMsg(msg proto.Message, f ServerMsgHandler) {
	p.client.processor.RegisterMsg(msg, f)
}

func (p *RPC) RegisterSessionMsg(msg proto.Message, f SessionMsgHandler) {
	p.client.processor.RegisterSessionMsg(msg, f)
}

func (p *RPC) RegisterRequest(msg proto.Message, f RequestHandler) {
	p.client.processor.RegisterRequest(msg, f)
}

func (p *RPC) GetServerById(serverID int32) Server {
	p.Lock()
	defer p.Unlock()
	if v, ok := p.sid2server[serverID]; ok {
		return v
	}
	s := NewServer(p.client, serverID)
	p.sid2server[serverID] = s
	return s
}

//TODO add
func (p *RPC) GetServerByType(serverType ServerType) Server {
	return nil
}

func (p *RPC) RegisterSend2Session(send2Session func(sesID int32, msgID uint32, data []byte)) {
	p.client.RegisterSend2Session(send2Session)
}
