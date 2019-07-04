package gate

import (
	"github.com/0990/goserver/network"
	"github.com/0990/goserver/rpc"
	"github.com/0990/goserver/service"
	"github.com/golang/protobuf/proto"
	"sync"
)

type Gate struct {
	WSAddr               string
	worker               service.Worker
	newEvent, closeEvent func(conn *Client)
	Processor            *Processor
	RPC                  *rpc.RPC

	sesid2Client map[int32]*Client
	sesMutex     sync.Mutex
}

func NewGate(addr string) (*Gate, error) {
	p := &Gate{
		WSAddr:    addr,
		Processor: NewProcessor(),
	}
	p.worker = service.NewWorker()
	rpc, err := rpc.NewRPC(100, p.worker)
	if err != nil {
		return nil, err
	}
	p.RPC = rpc
	p.RPC.RegisterSend2Session(func(sesid int32, msgid uint16, data []byte) {
		p.GetSession(sesid).SendRawMsg(msgid, data)
	})
	return p, nil
}

//TODO 添加关闭信号
func (p *Gate) Run() {
	//创建websocket server
	wss := network.NewWSServer(p.WSAddr, func(conn network.Conn) network.NewClienter {
		c := NewClient(conn, p)
		return c
	})
	wss.Start()
	p.worker.Run()
	p.RPC.Run()
}

func (p *Gate) Post(f func()) {
	p.worker.Post(f)
}

func (p *Gate) RegisterSessionEvent(new, close func(conn *Client)) {
	p.newEvent = new
	p.closeEvent = close
}

//路由服务
func (p *Gate) Router(msg interface{}) {

}

func (p *Gate) RegisterSessionHandler(msg proto.Message, f func(*Client, proto.Message)) {
	p.Processor.Register(msg)
	p.Processor.SetHandler(msg, f)
}

//func (p *Gate) RegisterRPCMsg(msg proto.Message, f func(sesid int32, msg proto.Message)) {
//	p.RPC.RegisterMsg(msg,f)
//}

func (p *Gate) RegisterRequestHandler(msg proto.Message, f func(rpc.Server, proto.Message)) {
	p.RPC.RegisterRequest(msg, f)
}

func (p *Gate) GetServerById(serverid int32) rpc.Server {
	return p.RPC.GetServerById(serverid)
}

func (p *Gate) GetSession(sesid int32) network.Session {
	p.sesMutex.Lock()
	defer p.sesMutex.Unlock()

	v, ok := p.sesid2Client[sesid]
	if !ok {
		return nil
	}
	return v
}
