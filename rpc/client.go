package rpc

import (
	"fmt"
	"github.com/0990/goserver/rpc/rpcmsg"
	"github.com/0990/goserver/service"
	"github.com/golang/protobuf/proto"
	"github.com/nats-io/nats.go"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"reflect"
	"time"
)

var ErrTimeOut = errors.New("rpc: timeout")
var ErrNoKnow = errors.New("rpc: unknow")

const CALL_TIMEOUT = 10 * time.Second

type Client struct {
	conn         *nats.Conn
	serverTopic  string
	serverID     int32
	send2Session func(sesID int32, msgID uint32, data []byte) //gate服专用
	processor    *Processor
	worker       service.Worker
	close        chan struct{}
}

func newClient(serverID int32, worker service.Worker, natsUrl string) (*Client, error) {
	p := &Client{}
	conn, err := nats.Connect(natsUrl)
	if err != nil {
		return nil, err
	}
	p.conn = conn
	p.serverID = serverID
	p.serverTopic = fmt.Sprintf("%v", serverID)
	p.processor = NewProcessor()
	p.worker = worker
	p.close = make(chan struct{})
	return p, nil
}

//阻塞式
func (p *Client) Call(serverTopic string, req proto.Message, resp proto.Message) error {
	ret := make(chan error)
	call := p.processor.RegisterCall(resp, func(err error) {
		ret <- err
	})

	data := MakeRequestData(req, call.seqID, p.serverID)
	err := p.publish(serverTopic, data)
	if err != nil {
		return err
	}

	select {
	case err, ok := <-ret:
		if !ok {
			return errors.New("client closed")
		}
		return err
	case <-time.After(CALL_TIMEOUT):
		p.processor.GetCallWithDel(call.seqID)
		return ErrTimeOut
	}

	return ErrNoKnow
}

//不需要注册Response的Request请求 onRecv func(*msg.XXX,error)
func (p *Client) Request(serverTopic string, msg proto.Message, cb interface{}) error {
	cbType := reflect.TypeOf(cb)
	if cbType.Kind() != reflect.Func {
		return errors.New("cb not a func")
	}
	cbValue := reflect.ValueOf(cb)
	numArgs := cbType.NumIn()
	if numArgs != 2 {
		return errors.New("cb param num args !=2")
	}
	args0 := cbType.In(0)
	if args0.Kind() != reflect.Ptr {
		return errors.New("cb param args0 not ptr")
	}
	//TODO 严格检查参数类型
	args1 := cbType.In(1)

	//TODO 如果出现error,resp==nil
	resp := reflect.New(args0.Elem()).Interface().(proto.Message)
	onRecv := func(err error) {
		oV := make([]reflect.Value, 2)
		oV[0] = reflect.ValueOf(resp)
		if err == nil {
			oV[1] = reflect.New(args1).Elem()
		} else {
			oV[1] = reflect.ValueOf(err)
		}
		cbValue.Call(oV)
	}

	call := p.processor.RegisterCall(resp, onRecv)
	data := MakeRequestData(msg, call.seqID, p.serverID)
	err := p.publish(serverTopic, data)
	if err != nil {
		return err
	}
	//util.PrintCurrNano("client request after")

	time.AfterFunc(CALL_TIMEOUT, func() {
		//TODO 放在主线程中工作
		if _, ok := p.processor.GetCallWithDel(call.seqID); ok {
			onRecv(ErrTimeOut)
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

func (p *Client) Close() (err error) {
	p.close <- struct{}{}
	return
}

func (p *Client) ReadLoop() error {
	sub, err := p.conn.SubscribeSync(p.serverTopic)
	if err != nil {
		return err
	}

	go func() {
		<-p.close
		sub.Unsubscribe()
	}()

	for {
		m, err := sub.NextMsg(time.Minute)
		if err != nil && err == nats.ErrTimeout {
			continue
		} else if err != nil {
			logrus.WithError(err).Error("ReadLoop NextMsg error")
			return err
		}
		rpcData := &rpcmsg.Data{}
		err = proto.Unmarshal(m.Data, rpcData)
		if err != nil {
			logrus.WithError(err)
			continue
		}
		p.worker.Post(func() {
			p.handle(rpcData)
		})
	}
}

func (p *Client) handle(rpcData *rpcmsg.Data) {
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
			return
		}
	case rpcmsg.Data_Response:
		err := p.processor.HandleResponse(seqID, data)
		if err != nil {
			logrus.WithError(err)
			return
		}
	case rpcmsg.Data_Session2Server:
		s := NewSession(p, senderID, sesID)
		err := p.processor.HandleSessionMsg(s, msgID, data)
		if err != nil {
			logrus.WithError(err)
			return
		}
	case rpcmsg.Data_Server2Session:
		p.send2Session(sesID, msgID, data)
	case rpcmsg.Data_Server2Server:
		s := NewServer(p, senderID)
		err := p.processor.HandleMsg(s, msgID, data)
		if err != nil {
			logrus.WithError(err)
			return
		}
	default:
		logrus.WithField("rpcType", rpcData.Type).Error("not support rpc type")
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
