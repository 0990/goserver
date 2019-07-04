package network

import "github.com/golang/protobuf/proto"

type NewClienter interface {
	ReadLoop()
	OnClose()
	OnNew()
}

type Session interface {
	SendMsg(msg proto.Message)
	SendRawMsg(msgid uint16, data []byte)
}
