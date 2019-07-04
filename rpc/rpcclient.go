package rpc

import (
	"fmt"
	"github.com/0990/golearn/rpc/netmsg"
	msg "github.com/0990/goserver/msg/rpc"
	"github.com/0990/goserver/util"
	"github.com/golang/protobuf/proto"
	"github.com/nats-io/go-nats"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"time"
)

var ErrTimeOut = errors.New("rpc: timeout")
var ErrNoKnow = errors.New("rpc: unknow")

type RPCClient struct {
	conn         *nats.Conn
	clientTopic  string
	serverid     int32
	send2Session func(sesid int32, msgid uint16, data []byte) //gate服专用
	processor    *Processor
}

func newRPCClient(serverid int32) (*RPCClient, error) {
	p := &RPCClient{}
	conn, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		return nil, err
	}
	p.conn = conn
	p.serverid = serverid
	p.clientTopic = fmt.Sprintf("%v", serverid)
	p.processor = NewProcessor()
	return p, nil
}

//阻塞式
func (p *RPCClient) Call(serverTopic string, message proto.Message) (proto.Message, error) {
	ret := make(chan proto.Message)
	call := CreateCall(func(msg proto.Message, err error) {
		ret <- msg
	})

	data := util.Marshal(message, seqid)

	err := p.send(serverTopic, data)
	if err != nil {
		return nil, err
	}

	select {
	case result, ok := <-ret:
		if !ok {
			return nil, errors.New("client closed")
		}
		return result, nil
	case <-time.After(time.Second * 10):
		GetCallWithDel(seqid)
		return nil, ErrTimeOut
	}

	return nil, ErrNoKnow
}

//非阻塞式
func (p *RPCClient) Request(serverTopic string, message proto.Message, onRecv func(proto.Message, error)) error {
	call := CreateCall(onRecv)
	data := util.Marshal(message, seqid)

	err := p.send(serverTopic, data)
	if err != nil {
		return err
	}

	time.AfterFunc(time.Second*10, func() {
		//TODO 放在主线程中工作
		if call, ok := GetCallWithDel(seqid); ok {
			onRecv(nil, ErrTimeOut)
		}
	})

	return nil
}

//仅发送
func (p *RPCClient) SendMsg(serverTopic string, msg proto.Message) {
	data := util.Marshal(msg, 0)
	p.send(serverTopic, data)
}

func (p *RPCClient) Answer(serverTopic string, seqid int32, msg proto.Message) {
	data := util.Marshal(msg, seqid)
	p.send(serverTopic, data)
}

//发送给gate，然后gate会发出去
func (p *RPCClient) RouteGate(gateTopic string, message proto.Message) {
	rpc := &msg.RPC{}
	rpc.Data, _ = proto.Marshal(message)
	rpc.Type = msg.RPC_RouteGate
	//TODO msgid
	rpc.Msgid = 0
	data, _ := proto.Marshal(rpc)
	p.send(gateTopic, data)
}

func (p *RPCClient) Run() {
	go p.ReadLoop()
}

func (p *RPCClient) ReadLoop() error {
	sub, err := p.conn.SubscribeSync(p.clientTopic)
	if err != nil {
		return err
	}

	for {
		m, err := sub.NextMsg(time.Minute)
		if err != nil && err == nats.ErrTimeout {
			continue
		} else if err != nil {
			return err
		}
		rpc := &msg.RPC{}
		err = proto.Unmarshal(m.Data, rpc)
		if err != nil {
			logrus.WithError(err)
			continue
		}
		msgid := uint16(rpc.Msgid)
		switch rpc.Type {
		case msg.RPC_Request:
			//rpc request
			s := NewRequestServer(p, rpc.Senderid, uint16(rpc.Sesid))
			err := p.processor.HandleRequest(s, msgid, rpc.Data)
			if err != nil {
				logrus.WithError(err)
				continue
			}
		case msg.RPC_Response:
			//rpc response
			if v, ok := GetCallWithDel(seqid); ok {
				onRecv(rpc, nil)
			}
		case msg.RPC_RouteServer:
			//route server
			//TODO unmarshal 构造func(session,fun)回调
			NewRPCSession(p, rpc.Senderid, rpc.Sesid)
		case msg.RPC_RouteGate:
			//route gate
			p.send2Session(rpc.Sesid, uint16(rpc.Msgid), rpc.Data)
		case msg.RPC_Normal:
			//normal
			s := NewServer(p, rpc.Senderid)
			err := p.processor.HandleMsg(s, msgid, rpc.Data)
			if err != nil {
				logrus.WithError(err)
				continue
			}
		default:
			logrus.WithField("rpcType", rpc.Type).Error("not support rpc type")
		}
	}
}

func (p *RPCClient) send(topic string, data []byte) error {
	return p.conn.Publish(topic, data)
}

func (p *RPCClient) Route2Server(topic string, sesid int32, message proto.Message) {
	data, _ := proto.Marshal(message)
	rpc := &msg.RPC{
		Type:     msg.RPC_RouteServer,
		Sesid:    sesid,
		Senderid: p.serverid,
		Msgid:    1,
		Data:     data,
	}

	rpcData, _ := proto.Marshal(rpc)
	p.send(topic, rpcData)
}

func (p *RPCClient) RegisterSend2Session(send2Session func(sesid int32, msgid uint16, data []byte)) {
	p.send2Session = send2Session
}
