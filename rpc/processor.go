package rpc

import (
	"errors"
	"github.com/0990/goserver/util"
	"github.com/golang/protobuf/proto"
	"github.com/sirupsen/logrus"
	"reflect"
)

type Processor struct {
	msgID2Request map[uint16]*RequestInfo
	msgID2Msg     map[uint16]*MsgInfo
}

type RequestInfo struct {
	msgType    reflect.Type
	msgHandler RequestHandler
}

type MsgInfo struct {
	msgType    reflect.Type
	msgHandler MsgHandler
}

type RequestHandler func(RequestServer, msg proto.Message)

type MsgHandler func(Server, msg proto.Message)

func NewProcessor() *Processor {
	p := new(Processor)
	p.msgID2Request = make(map[uint16]*RequestInfo)
	p.msgID2Msg = make(map[uint16]*MsgInfo)
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

func (p *Processor) RegisterMsg(msg proto.Message, f MsgHandler) {
	msgID, msgType := util.ProtoHash(msg)
	if _, ok := p.msgID2Request[msgID]; ok {
		logrus.Errorf("message %s is already registered", msgType)
		return
	}

	msgInfo := new(MsgInfo)
	msgInfo.msgType = msgType
	msgInfo.msgHandler = f
	p.msgID2Msg[msgID] = msgInfo
}

func (p *Processor) decodeRequest(msgid uint16, data []byte) (proto.Message, error) {
	msgInfo, exist := p.msgID2Request[msgid]
	if !exist {
		return nil, errors.New("msgID not registered")
	}

	msg := reflect.New(msgInfo.msgType.Elem()).Interface().(proto.Message)
	return msg, proto.Unmarshal(data, msg.(proto.Message))
}

func (p *Processor) DecodeMsg(msgid uint16, data []byte) (proto.Message, error) {
	msgInfo, exist := p.msgID2Msg[msgid]
	if !exist {
		return nil, errors.New("msgID not registered")
	}

	msg := reflect.New(msgInfo.msgType.Elem()).Interface().(proto.Message)
	return msg, proto.Unmarshal(data, msg.(proto.Message))
}

func (p *Processor) HandleRequest(server RequestServer, msgid uint16, data []byte) error {
	msgInfo, ok := p.msgID2Request[msgid]
	if !ok {
		logrus.Errorf("message %s not registered", msgid)
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

func (p *Processor) HandleMsg(server Server, msgid uint16, data []byte) error {
	msgInfo, ok := p.msgID2Msg[msgid]
	if !ok {
		logrus.Errorf("message %s not registered", msgid)
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
