package network

import (
	"github.com/0990/goserver/service"
	"github.com/golang/protobuf/proto"
	"sync"
)

type Mgr struct {
	wsAddr         string
	sesID2Client   map[int32]*Client
	sesMutex       sync.Mutex
	onNew, onClose func(conn Session)
	processor      *Processor
	worker         service.Worker
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

func (p *Mgr) Run() error {
	//创建websocket server
	wss := NewWSServer(p.wsAddr, func(conn Conn) *Client {
		c := NewClient(conn, p)
		return c
	})
	wss.Start()
	return nil
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

func (p *Mgr) RegisterSessionHandler(msg proto.Message, f func(Session, proto.Message)) {
	p.processor.Register(msg)
	p.processor.SetHandler(msg, f)
}
