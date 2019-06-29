package network

import (
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"sync"
)

//这里可以定义handshake规则：超时时间等
var upgrader = websocket.Upgrader{} // use default options

type WSServer struct {
	sync.Mutex
	addr  string
	wg    sync.WaitGroup
	conns map[*websocket.Conn]struct{}

	newClient func(conn Conn) Clienter
}

func NewWSServer(addr string, newClient func(conn Conn) Clienter) *WSServer {
	return &WSServer{
		addr:      addr,
		newClient: newClient,
		conns:     make(map[*websocket.Conn]struct{}),
	}
}

func (p *WSServer) Start() {
	//TODO 监听失败情况下，要中止程序
	go func() {
		err := http.ListenAndServe(p.addr, p)
		if err != nil {
			log.Fatal("ListenAndServe: ", err)
		}
	}()
}

func (p *WSServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	p.wg.Add(1)
	defer p.wg.Done()

	p.Lock()
	p.conns[conn] = struct{}{}
	p.Unlock()

	wsc := NewWSConn(conn)
	go wsc.writePump()

	c := p.newClient(wsc)
	c.ReadLoop()

	wsc.Close()
	p.Lock()
	delete(p.conns, conn)
	p.Unlock()
	c.OnClose()
}
