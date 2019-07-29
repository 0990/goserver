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
	Notify(proto.Message)
	Call(proto.Message, proto.Message) error
	RouteSession2Server(sesID int32, msg proto.Message)

	Request(proto.Message, interface{}) error

	ID() int32
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

func (p *server) Notify(msg proto.Message) {
	p.rpcClient.SendMsg(p.serverTopic, msg)
}

func (p *server) ID() int32 {
	return p.serverid
}

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

func (p *requestserver) ID() int32 {
	return p.serverid
}

type Session interface {
	SendMsg(msg proto.Message)
	SendRawMsg(msgID uint16, data []byte)
	GateSessionID() GateSessionID
}

type GateSessionID struct {
	GateID int32
	SesID  int32
}

type session struct {
	gsID      GateSessionID
	rpcClient *Client
	gateTopic string
}

func NewSession(client *Client, gateID int32, sesID int32) Session {
	g := GateSessionID{GateID: gateID, SesID: sesID}

	return &session{
		gsID:      g,
		rpcClient: client,
		gateTopic: fmt.Sprintf("%v", gateID),
	}
}

func (p *session) SendMsg(msg proto.Message) {
	p.rpcClient.RouteGate(p.gateTopic, p.gsID.SesID, msg)
}

func (p *session) SendRawMsg(msgID uint16, data []byte) {

}

func (p *session) GateSessionID() GateSessionID {
	return p.gsID
}
