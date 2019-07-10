package rpc

import (
	"errors"
	"github.com/0990/goserver/util"
	"github.com/golang/protobuf/proto"
	"github.com/sirupsen/logrus"
	"reflect"
	"sync"
	"sync/atomic"
)

type Processor struct {
	msgID2Request    map[uint32]*RequestInfo
	msgID2ServerMsg  map[uint32]*ServerMsgInfo
	msgID2SessionMsg map[uint32]*SessionMsgInfo

	seqID          int32
	seqID2CallInfo sync.Map
}

type RequestInfo struct {
	msgType    reflect.Type
	msgHandler RequestHandler
}

type ServerMsgInfo struct {
	msgType    reflect.Type
	msgHandler ServerMsgHandler
}

type SessionMsgInfo struct {
	msgType    reflect.Type
	msgHandler SessionMsgHandler
}

type RequestHandler func(RequestServer, proto.Message)
type ServerMsgHandler func(Server, proto.Message)
type SessionMsgHandler func(Session, proto.Message)

func NewProcessor() *Processor {
	p := new(Processor)
	p.msgID2Request = make(map[uint32]*RequestInfo)
	p.msgID2ServerMsg = make(map[uint32]*ServerMsgInfo)
	p.msgID2SessionMsg = make(map[uint32]*SessionMsgInfo)

	return p
}

func (p *Processor) RegisterRequest(msg proto.Message, f RequestHandler) {
	msgID, msgType := util.ProtoHash(msg)
	if _, ok := p.msgID2Request[msgID]; ok {
		logrus.Errorf("message %s is already registered", msgType)
		return
	}

	msgInfo := new(RequestInfo)
	msgInfo.msgType = msgType
	msgInfo.msgHandler = f
	p.msgID2Request[msgID] = msgInfo
}

func (p *Processor) RegisterMsg(msg proto.Message, f ServerMsgHandler) {
	msgID, msgType := util.ProtoHash(msg)
	if _, ok := p.msgID2ServerMsg[msgID]; ok {
		logrus.Errorf("message %s is already registered", msgType)
		return
	}

	msgInfo := new(ServerMsgInfo)
	msgInfo.msgType = msgType
	msgInfo.msgHandler = f
	p.msgID2ServerMsg[msgID] = msgInfo
}

func (p *Processor) RegisterSessionMsg(msg proto.Message, f SessionMsgHandler) {
	msgID, msgType := util.ProtoHash(msg)
	if _, ok := p.msgID2SessionMsg[msgID]; ok {
		logrus.Errorf("message %s is already registered", msgType)
		return
	}

	msgInfo := new(SessionMsgInfo)
	msgInfo.msgType = msgType
	msgInfo.msgHandler = f
	p.msgID2SessionMsg[msgID] = msgInfo
}

func (p *Processor) HandleRequest(server RequestServer, msgID uint32, data []byte) error {
	msgInfo, ok := p.msgID2Request[msgID]
	if !ok {
		logrus.Errorf("message %s not registered", msgID)
		return errors.New("msgID not registered")
	}

	msg := reflect.New(msgInfo.msgType.Elem()).Interface().(proto.Message)
	err := proto.Unmarshal(data, msg)
	if err != nil {
		logrus.WithError(err).Error("HandleRequest")
		return err
	}
	if msgInfo.msgHandler != nil {
		msgInfo.msgHandler(server, msg)
	}
	return nil
}

func (p *Processor) HandleMsg(server Server, msgID uint32, data []byte) error {
	msgInfo, ok := p.msgID2ServerMsg[msgID]
	if !ok {
		logrus.Errorf("message %s not registered", msgID)
		return errors.New("msgID not registered")
	}

	msg := reflect.New(msgInfo.msgType.Elem()).Interface().(proto.Message)
	err := proto.Unmarshal(data, msg)
	if err != nil {
		logrus.WithError(err).Error("HandleRequest")
		return err
	}
	if msgInfo.msgHandler != nil {
		msgInfo.msgHandler(server, msg)
	}
	return nil
}

func (p *Processor) HandleSessionMsg(session Session, msgID uint32, data []byte) error {
	msgInfo, ok := p.msgID2SessionMsg[msgID]
	if !ok {
		logrus.Errorf("message %s not registered", msgID)
		return errors.New("msgID not registered")
	}

	msg := reflect.New(msgInfo.msgType.Elem()).Interface().(proto.Message)
	err := proto.Unmarshal(data, msg)
	if err != nil {
		logrus.WithError(err).Error("HandleRequest")
		return err
	}
	if msgInfo.msgHandler != nil {
		msgInfo.msgHandler(session, msg)
	}
	return nil
}

func (p *Processor) NewSeqID() int32 {
	return atomic.AddInt32(&p.seqID, 1)
}

type Call struct {
	seqID  int32
	onRecv func(error)
	resp   proto.Message
}

func (p *Processor) RegisterCall(resp proto.Message, onRecv func(error)) *Call {
	seqID := p.NewSeqID()
	call := &Call{
		seqID:  seqID,
		onRecv: onRecv,
		resp:   resp,
	}
	p.seqID2CallInfo.Store(seqID, call)
	return call
}

func (p *Processor) GetCallWithDel(seqID int32) (*Call, bool) {
	if v, ok := p.seqID2CallInfo.Load(seqID); ok {
		p.seqID2CallInfo.Delete(seqID)
		return v.(*Call), true
	}
	return nil, false
}

func (p *Processor) HandleResponse(seqID int32, data []byte) error {
	call, ok := p.GetCallWithDel(seqID)
	if !ok {
		return errors.New("seqID not existed")
	}
	err := proto.Unmarshal(data, call.resp)
	if err != nil {
		logrus.WithError(err).Error("HandleRequest")
		call.onRecv(err)
		return err
	}

	call.onRecv(nil)
	return nil
}
