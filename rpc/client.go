package rpc

import (
	"fmt"
	"github.com/0990/goserver/rpc/rpcmsg"
	"github.com/golang/protobuf/proto"
	"github.com/nats-io/go-nats"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"time"
)

var ErrTimeOut = errors.New("rpc: timeout")
var ErrNoKnow = errors.New("rpc: unknow")

type Client struct {
	conn         *nats.Conn
	serverTopic  string
	serverID     int32
	send2Session func(sesID int32, msgID uint32, data []byte) //gate服专用
	processor    *Processor
}

func newClient(serverID int32) (*Client, error) {
	p := &Client{}
	conn, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		return nil, err
	}
	p.conn = conn
	p.serverID = serverID
	p.serverTopic = fmt.Sprintf("%v", serverID)
	p.processor = NewProcessor()
	return p, nil
}

//阻塞式
func (p *Client) Call(serverTopic string, msg proto.Message) (proto.Message, error) {
	ret := make(chan proto.Message)

	call := p.processor.RegisterCall(func(msg proto.Message, err error) {
		ret <- msg
	})

	data := MakeRequestData(msg, call.seqID, p.serverID)
	err := p.publish(serverTopic, data)
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
		p.processor.GetCallWithDel(call.seqID)
		return nil, ErrTimeOut
	}

	return nil, ErrNoKnow
}

//非阻塞式
func (p *Client) Request(serverTopic string, msg proto.Message, onRecv func(proto.Message, error)) error {
	call := p.processor.RegisterCall(onRecv)
	data := MakeRequestData(msg, call.seqID, p.serverID)
	err := p.publish(serverTopic, data)
	if err != nil {
		return err
	}

	time.AfterFunc(time.Second*10, func() {
		//TODO 放在主线程中工作
		if _, ok := p.processor.GetCallWithDel(call.seqID); ok {
			onRecv(nil, ErrTimeOut)
		}
	})

	return nil
}

//仅发送
func (p *Client) SendMsg(serverTopic string, msg proto.Message) {
	data := MakeServer2ServerData(msg, p.serverID)
	p.publish(serverTopic, data)
}

func (p *Client) Answer(serverTopic string, seqid int32, msg proto.Message) {
	data := MakeResponseData(msg, seqid, p.serverID)
	p.publish(serverTopic, data)
}

//发送给gate，然后gate会发出去
func (p *Client) RouteGate(gateTopic string, sesID int32, msg proto.Message) {
	data := MakeServer2SessionData(msg, sesID, p.serverID)
	p.publish(gateTopic, data)
}

func (p *Client) Run() {
	go p.ReadLoop()
}

func (p *Client) ReadLoop() error {
	sub, err := p.conn.SubscribeSync(p.serverTopic)
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
		rpcData := &rpcmsg.Data{}
		err = proto.Unmarshal(m.Data, rpcData)
		if err != nil {
			logrus.WithError(err)
			continue
		}
		msgID := rpcData.Msgid
		seqID := rpcData.Seqid
		sesID := rpcData.Sesid
		senderID := rpcData.Senderid
		data := rpcData.Data

		switch rpcData.Type {
		case rpcmsg.Data_Request:
			s := NewRequestServer(p, senderID, seqID)
			err := p.processor.HandleRequest(s, msgID, data)
			if err != nil {
				logrus.WithError(err)
				continue
			}
		case rpcmsg.Data_Response:
			err := p.processor.HandleResponse(seqID, msgID, data)
			if err != nil {
				logrus.WithError(err)
				continue
			}
		case rpcmsg.Data_Session2Server:
			s := NewSession(p, senderID, sesID)
			err := p.processor.HandleSessionMsg(s, msgID, data)
			if err != nil {
				logrus.WithError(err)
				continue
			}
		case rpcmsg.Data_Server2Session:
			p.send2Session(sesID, msgID, data)
		case rpcmsg.Data_Server2Server:
			s := NewServer(p, senderID)
			err := p.processor.HandleMsg(s, msgID, data)
			if err != nil {
				logrus.WithError(err)
				continue
			}
		default:
			logrus.WithField("rpcType", rpcData.Type).Error("not support rpc type")
		}
	}
}

func (p *Client) publish(topic string, data []byte) error {
	return p.conn.Publish(topic, data)
}

func (p *Client) RouteSession2Server(topic string, sesID int32, msg proto.Message) {
	data := MakeSession2ServerData(msg, sesID, p.serverID)
	p.publish(topic, data)
}

func (p *Client) RegisterSend2Session(send2Session func(sesID int32, msgID uint32, data []byte)) {
	p.send2Session = send2Session
}
