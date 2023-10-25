package network

import (
	"github.com/0990/goserver/service"
	"github.com/0990/goserver/util"
	"github.com/golang/protobuf/proto"
	"github.com/sirupsen/logrus"
	"net"
	"reflect"
	"sync"
)

type Mgr struct {
	wsAddr         string
	sesID2Client   map[int32]*Client
	sesMutex       sync.Mutex
	onNew, onClose func(conn Session)
	processor      *Processor
	worker         service.Worker
	wss            *WSServer

	close func()
}

func NewMgr(wsAddr string, worker service.Worker) *Mgr {
	p := &Mgr{
		wsAddr:       wsAddr,
		sesID2Client: make(map[int32]*Client),
		worker:       worker,
		processor:    NewProcessor(),
	}
	return p
}

//func (p *Mgr) Init(wsAddr string, worker service.Worker) {
//	p.wsAddr = wsAddr
//	p.sesID2Client = make(map[int32]*Client)
//	p.worker = worker
//}

func (p *Mgr) Run() {
	//创建websocket server
	wss := NewWSServer(p.wsAddr, func(conn Conn) *Client {
		c := NewClient(conn, p)
		return c
	})
	p.wss = wss
	wss.Start()
	p.close = wss.Close
}

func (p *Mgr) ListenAddr() *net.TCPAddr {
	return p.wss.ListenAddr()
}

func (p *Mgr) Close() {
	p.close()
}

func (p *Mgr) Post(f func()) {
	p.worker.Post(f)
}

func (p *Mgr) RegisterEvent(onNew, onClose func(conn Session)) {
	p.onNew = onNew
	p.onClose = onClose
}

func (p *Mgr) GetSession(sesID int32) (Session, bool) {
	p.sesMutex.Lock()
	defer p.sesMutex.Unlock()

	v, ok := p.sesID2Client[sesID]
	return v, ok
}

//func (p *Mgr) RegisterSessionMsgHandler(msg proto.Message, f func(Session, proto.Message)) {
//	p.processor.Register(msg)
//	p.processor.SetHandler(msg, f)
//}

func (p *Mgr) RegisterSessionMsgHandler(cb interface{}) {
	err, funValue, msgType := util.CheckArgs1MsgFun(cb)
	if err != nil {
		logrus.WithError(err).Error("RegisterServerMsgHandler")
		return
	}
	msg := reflect.New(msgType).Elem().Interface().(proto.Message)
	p.processor.RegisterSessionMsgHandler(msg, func(s Session, message proto.Message) {
		funValue.Call([]reflect.Value{reflect.ValueOf(s), reflect.ValueOf(message)})
	})
}

func (p *Mgr) RegisterRawSessionMsgHandler(msg proto.Message, handler MsgHandler) {
	p.processor.RegisterSessionMsgHandler(msg, handler)
}
