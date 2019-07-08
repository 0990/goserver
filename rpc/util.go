package rpc

import (
	"github.com/0990/goserver/rpc/rpcmsg"
	"github.com/0990/goserver/util"
	"github.com/golang/protobuf/proto"
)

func MakeRequestData(msg proto.Message, seqID int32, senderID int32) []byte {
	msgID, _ := util.ProtoHash(msg)
	msgData, _ := proto.Marshal(msg)
	rpc := &rpcmsg.Data{
		Type:     rpcmsg.Data_Request,
		Msgid:    msgID,
		Seqid:    seqID,
		Senderid: senderID,
		Data:     msgData,
	}

	data, _ := proto.Marshal(rpc)
	return data
}

func MakeServer2ServerData(msg proto.Message, senderID int32) []byte {
	msgID, _ := util.ProtoHash(msg)
	msgData, _ := proto.Marshal(msg)
	rpc := &rpcmsg.Data{
		Type:     rpcmsg.Data_Server2Server,
		Msgid:    msgID,
		Senderid: senderID,
		Data:     msgData,
	}

	data, _ := proto.Marshal(rpc)
	return data
}

func MakeResponseData(msg proto.Message, seqID int32, senderID int32) []byte {
	msgID, _ := util.ProtoHash(msg)
	msgData, _ := proto.Marshal(msg)
	rpc := &rpcmsg.Data{
		Type:     rpcmsg.Data_Response,
		Msgid:    msgID,
		Senderid: senderID,
		Data:     msgData,
	}

	data, _ := proto.Marshal(rpc)
	return data
}

func MakeServer2SessionData(msg proto.Message, sesID int32, senderID int32) []byte {
	msgID, _ := util.ProtoHash(msg)
	msgData, _ := proto.Marshal(msg)
	rpc := &rpcmsg.Data{
		Type:     rpcmsg.Data_Server2Session,
		Msgid:    msgID,
		Senderid: senderID,
		Data:     msgData,
		Sesid:    sesID,
	}

	data, _ := proto.Marshal(rpc)
	return data
}

func MakeSession2ServerData(msg proto.Message, sesID int32, senderID int32) []byte {
	msgID, _ := util.ProtoHash(msg)
	msgData, _ := proto.Marshal(msg)
	rpc := &rpcmsg.Data{
		Type:     rpcmsg.Data_Session2Server,
		Msgid:    msgID,
		Senderid: senderID,
		Data:     msgData,
		Sesid:    sesID,
	}

	data, _ := proto.Marshal(rpc)
	return data
}
