package rpc

import (
	"fmt"
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
	Call(proto.Message, proto.Message) error
	RouteSession2Server(sesID int32, msg proto.Message)

	//func(*msg.XXX, error))
	Request(proto.Message, interface{}) error
}

type server struct {
	rpcClient   *Client
	serverid    int32
	serverTopic string //目标服务器nats的topic,暂为服务器id
}

func NewServer(client *Client, serverid int32) Server {
	return &server{
		rpcClient:   client,
		serverid:    serverid,
		serverTopic: fmt.Sprintf("%v", serverid),
	}
}

func (p *server) Send(msg proto.Message) {
	p.rpcClient.SendMsg(p.serverTopic, msg)
}

//func (p *server) Request(msg proto.Message, f func(proto.Message, error)) error {
//	p.rpcClient.Request(p.serverTopic, msg, f)
//	return nil
//}

//func (p *server) Request(msg proto.Message, f func(proto.Message, error)) error {
//	p.rpcClient.Request(p.serverTopic, msg, f)
//	return nil
//}

func (p *server) Request(msg proto.Message, f interface{}) error {
	p.rpcClient.Request(p.serverTopic, msg, f)
	return nil
}

func (p *server) Call(req proto.Message, resp proto.Message) error {
	return p.rpcClient.Call(p.serverTopic, req, resp)
}

//gate服使用较多，把消息路由到对应服务器
func (p *server) RouteSession2Server(sesid int32, msg proto.Message) {
	p.rpcClient.RouteSession2Server(p.serverTopic, sesid, msg)
}

type RequestServer interface {
	Answer(proto.Message)
	Server
}

func NewRequestServer(client *Client, serverid int32, seqid int32) RequestServer {
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

type Session interface {
	SendMsg(msg proto.Message)
	SendRawMsg(msgID uint16, data []byte)
}

type session struct {
	sid       int32
	rpcClient *Client
	gateID    int32
	gateTopic string
}

func NewSession(client *Client, gateID int32, sesID int32) Session {
	return &session{
		sid:       sesID,
		gateID:    gateID,
		rpcClient: client,
		gateTopic: fmt.Sprintf("%v", gateID),
	}
}

func (p *session) SendMsg(msg proto.Message) {
	p.rpcClient.RouteGate(p.gateTopic, p.sid, msg)
}

func (p *session) SendRawMsg(msgID uint16, data []byte) {

}
