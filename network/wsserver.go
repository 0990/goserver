package network

import (
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"net"
	"net/http"
	"sync"
	"sync/atomic"
)

//这里可以定义handshake规则：超时时间等
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
	Subprotocols: []string{"avatar-fight"},
} // use default options

type WSServer struct {
	sync.Mutex
	addr      string
	wg        sync.WaitGroup
	conns     map[*websocket.Conn]struct{}
	connid    int32
	newClient func(conn Conn) *Client
}

func NewWSServer(addr string, newClient func(conn Conn) *Client) *WSServer {
	return &WSServer{
		addr:      addr,
		newClient: newClient,
		conns:     make(map[*websocket.Conn]struct{}),
	}
}

func (p *WSServer) Start() error {
	ln, err := net.Listen("tcp", p.addr)
	if err != nil {
		logrus.WithField("addr", p.addr).Fatal("启动失败，端口被占用")
		return err
	}
	go func() {
		err := http.Serve(ln, p)
		if err != nil {
			logrus.WithError(err).Fatal("WSServer Serve")
			return
		}
	}()
	return nil
}

func (p *WSServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logrus.WithError(err).Error("ServerHttp upgrader.Upgrade")
		return
	}

	p.wg.Add(1)
	defer p.wg.Done()

	p.Lock()
	p.conns[conn] = struct{}{}
	p.Unlock()

	connid := p.NewConnID()
	wsc := NewWSConn(conn, connid)
	go wsc.writePump()

	c := p.newClient(wsc)
	c.OnNew()
	c.ReadLoop()

	wsc.Close()
	p.Lock()
	delete(p.conns, conn)
	p.Unlock()
	c.OnClose()
}

func (p *WSServer) NewConnID() int32 {
	return atomic.AddInt32(&p.connid, 1)
}
