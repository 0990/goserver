package network

import (
	"encoding/binary"
	"github.com/0990/goserver/util"
	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"reflect"
)

type Processor struct {
	littleEndian bool
	msgID2Info   map[uint32]*MsgInfo
}

type MsgInfo struct {
	msgType    reflect.Type
	msgHandler MsgHandler
}

type MsgHandler func(client Session, msg proto.Message)

func NewProcessor() *Processor {
	p := new(Processor)
	p.littleEndian = false
	p.msgID2Info = make(map[uint32]*MsgInfo)
	return p
}

func (p *Processor) SetByteOrder(littleEndian bool) {
	p.littleEndian = littleEndian
}

func (p *Processor) Register(msg proto.Message) {
	msgID, msgType := util.ProtoHash(msg)
	if _, ok := p.msgID2Info[msgID]; ok {
		logrus.Errorf("message %s is already registered", msgType)
		return
	}

	msgInfo := new(MsgInfo)
	msgInfo.msgType = msgType
	p.msgID2Info[msgID] = msgInfo
	return
}

func (p *Processor) SetHandler(msg proto.Message, msgHandler MsgHandler) {
	msgID, msgType := util.ProtoHash(msg)
	msgInfo, ok := p.msgID2Info[msgID]
	if !ok {
		logrus.Errorf("message %s not registered", msgType)
		return
	}

	msgInfo.msgHandler = msgHandler
}

func (p *Processor) Handle(msg proto.Message, client *Client) error {
	msgID, msgType := util.ProtoHash(msg)
	msgInfo, ok := p.msgID2Info[msgID]
	if !ok {
		logrus.Errorf("message %s not registered", msgType)
		return nil
	}

	if msgInfo.msgHandler != nil {
		msgInfo.msgHandler(client, msg)
	}

	return nil
}

func (p *Processor) Unmarshal(data []byte) (proto.Message, error) {
	if len(data) < 2 {
		return nil, errors.New("protobuf data too short")
	}

	var msgID uint32
	if p.littleEndian {
		msgID = binary.LittleEndian.Uint32(data)
	} else {
		msgID = binary.BigEndian.Uint32(data)
	}

	msgInfo, exist := p.msgID2Info[msgID]
	if !exist {
		return nil, errors.New("msgID not registered")
	}

	msg := reflect.New(msgInfo.msgType.Elem()).Interface().(proto.Message)
	return msg, proto.Unmarshal(data[2:], msg.(proto.Message))
}

func (p *Processor) Marshal(msg proto.Message) ([]byte, error) {
	msgID, _ := util.ProtoHash(msg)

	data, err := proto.Marshal(msg)
	if err != nil {
		return nil, err
	}
	return p.Encode(msgID, data), nil
	//msgIDData := make([]byte, 2)
	//if p.littleEndian {
	//	binary.LittleEndian.PutUint16(msgIDData, msgID)
	//} else {
	//	binary.BigEndian.PutUint16(msgIDData, msgID)
	//}
	//
	////TODO 性能优化
	//ret := append(msgIDData, data...)
	//return ret, nil
}

func (p *Processor) Encode(msgid uint32, data []byte) []byte {
	msgIDData := make([]byte, 4)
	if p.littleEndian {
		binary.LittleEndian.PutUint32(msgIDData, msgid)
	} else {
		binary.BigEndian.PutUint32(msgIDData, msgid)
	}

	//TODO 性能优化
	ret := append(msgIDData, data...)
	return ret
}
