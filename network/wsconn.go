package network

import (
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
		return nil
	}

	return p.doWrite(args)

	//var msgLen uint32
	//for i := 0; i < len(args); i++ {
	//	msgLen += uint32(len(args[i]))
	//}
	//
	//if len(args) == 1 {
	//	return p.doWrite(args[0])
	//}
	//
	//msg := make([]byte, msgLen)
	//l := 0
	//for i := 0; i < len(args); i++ {
	//	copy(msg[l:], args[i])
	//	l += len(args[i])
	//}
	//return p.doWrite(msg)
}

func (p *WSConn) doWrite(data []byte) error {
	//TODO send chan 堵满情况处理
	p.send <- data
	return nil
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
