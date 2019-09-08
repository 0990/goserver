package network

import (
	"errors"
	"github.com/gorilla/websocket"
	"net"
	"sync"
)

type WSConn struct {
	sync.Mutex
	conn *websocket.Conn

	// Buffered channel of outbound messages.
	send chan []byte

	closeFlag bool

	connid int32
}

func NewWSConn(conn *websocket.Conn, connid int32) *WSConn {
	wsc := new(WSConn)
	wsc.send = make(chan []byte, 256)
	wsc.conn = conn
	wsc.connid = connid
	return wsc
}

func (p *WSConn) writePump() {
	for data := range p.send {
		if data == nil {
			break
		}
		err := p.conn.WriteMessage(websocket.BinaryMessage, data)
		if err != nil {
			break
		}
	}
	p.conn.Close()
	p.Lock()
	p.closeFlag = true
	p.Unlock()
}

func (p *WSConn) WriteMsg(args []byte) error {
	p.Lock()
	defer p.Unlock()

	if p.closeFlag {
		return errors.New("socket closeFlag is true")
	}

	if len(p.send) == cap(p.send) {
		close(p.send)
		p.closeFlag = true
		return errors.New("send buffer full")
	}

	p.send <- args
	return nil
}

func (p *WSConn) doWrite(data []byte) {
	p.send <- data
}

func (p *WSConn) ReadMsg() ([]byte, error) {
	_, data, err := p.conn.ReadMessage()
	return data, err
}

func (p *WSConn) ID() int32 {
	return p.connid
}

func (p *WSConn) Close() {
	p.Lock()
	defer p.Unlock()
	if p.closeFlag {
		return
	}
	p.doWrite(nil)
	p.closeFlag = true
}

func (p *WSConn) LocalAddr() net.Addr {
	return p.conn.LocalAddr()
}

func (p *WSConn) RemoteAddr() net.Addr {
	return p.conn.RemoteAddr()
}
