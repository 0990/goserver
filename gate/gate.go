package gate

import "github.com/0990/goserver/network"

type Gate struct {
	WSAddr               string
	workChan             chan func()
	newEvent, closeEvent func(conn *Client)
}

func NewGate(addr string) *Gate {
	return &Gate{
		WSAddr:   addr,
		workChan: make(chan func(), 1024),
	}
}

//TODO 添加关闭信号
func (p *Gate) Run() {
	wss := network.NewWSServer(p.WSAddr, func(conn network.Conn) network.Clienter {
		c := NewClient(conn, p)
		p.Post(func() {
			p.newEvent(c)
		})
		return c
	})
	wss.Start()
	p.WorkRun()
}

func (p *Gate) WorkRun() {
	go func() {
		for v := range p.workChan {
			v()
		}
	}()
}

func (p *Gate) Post(f func()) {
	p.workChan <- f
}

func (p *Gate) RegisterSessionEvent(new, close func(conn *Client)) {
	p.newEvent = new
	p.closeEvent = close
}
