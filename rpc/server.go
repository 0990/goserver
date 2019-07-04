package rpc

import (
	"fmt"
	"github.com/0990/goserver/network"
	"github.com/golang/protobuf/proto"
)

type ServerType int

const (
	_ ServerType = iota
	Gate
	Center
	Game
	Users
)

type Server interface {
	Send(proto.Message)
	Request(proto.Message, func(proto.Message, error)) error
	Call(proto.Message) (proto.Message, error)
	Route(sesid int32, msg proto.Message)
}

type server struct {
	rpcClient   *RPCClient
	serverid    int32
	serverTopic string //目标服务器nats的topic,暂为服务器id
}

func NewServer(client *RPCClient, serverid int32) Server {
	return &server{
		rpcClient:   client,
		serverid:    serverid,
		serverTopic: fmt.Sprintf("%v", serverid),
	}
}

func (p *server) Send(msg proto.Message) {
	p.rpcClient.SendMsg(p.serverTopic, msg)
}

func (p *server) Request(msg proto.Message, f func(proto.Message, error)) error {
	p.rpcClient.Request(p.serverTopic, msg, f)
	return nil
}

func (p *server) Call(msg proto.Message) (proto.Message, error) {
	return p.rpcClient.Call(p.serverTopic, msg)
}

func (p *server) Route(sesid int32, msg proto.Message) {
	p.rpcClient.Route2Server(p.serverTopic, sesid, msg)
}

type RequestServer interface {
	Answer(proto.Message)
	Server
}

func NewRequestServer(client *RPCClient, serverid int32, seqid int32) RequestServer {
	s := &server{
		rpcClient:   client,
		serverid:    serverid,
		serverTopic: fmt.Sprintf("%v", serverid),
	}
	return &requestserver{
		server: s,
		seqid:  seqid,
	}
}

type requestserver struct {
	*server
	seqid int32
}

func (p *requestserver) Answer(msg proto.Message) {
	p.server.rpcClient.Answer(p.serverTopic, p.seqid, msg)
}

type RPCSession struct {
	sesid        int32
	rpcClient    *RPCClient
	gateserverid int32
	gateTopic    string
}

func NewRPCSession(client *RPCClient, gateserverid int32, sesid int32) network.Session {
	return &RPCSession{
		sesid:        sesid,
		gateserverid: gateserverid,
		rpcClient:    client,
		gateTopic:    fmt.Sprintf("%v", gateserverid),
	}
}

func (p *RPCSession) SendMsg(msg proto.Message) {
	p.rpcClient.RouteGate(p.gateTopic, msg)
}

func (p *RPCSession) SendRawMsg(msgid uint16, data []byte) {

}
