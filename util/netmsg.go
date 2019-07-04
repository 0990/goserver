package util

import (
	msg "github.com/0990/goserver/msg/rpc"
	"github.com/golang/protobuf/proto"
)

func Marshal(message proto.Message, seqid int32) []byte {
	rpc := &msg.RPC{}
	rpc.Data, _ = proto.Marshal(message)
	rpc.Seqid = seqid
	rpc.Type = msg.RPC_Request
	//TODO msgid
	rpc.Msgid = 0
	data, _ := proto.Marshal(rpc)
	return data
}
