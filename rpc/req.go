package rpc

import (
	"github.com/0990/golearn/rpc/msg"
	"github.com/golang/protobuf/proto"
	"sync"
	"sync/atomic"
)

var (
	seqid      int32
	seqid2Call sync.Map
)

type Call struct {
	seqid  int32
	onRecv func(*msg.RPC, error)
}

func IncrSeqid() int32 {
	return atomic.AddInt32(&seqid, 1)
}

func CreateCall(onRecv func(proto.Message, error)) *Call {
	seqid := IncrSeqid()
	call := &Call{
		seqid:  seqid,
		onRecv: onRecv,
	}
	seqid2Call.Store(seqid, call)
	return call
}

func GetCallWithDel(seqid int32) (*Call, bool) {
	if v, ok := seqid2Call.Load(seqid); ok {
		seqid2Call.Delete(seqid)
		return v.(*Call), true
	}
	return nil, false
}
