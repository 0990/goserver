package gate

import (
	"fmt"
	"github.com/0990/goserver/network"
	"reflect"
)

type Client struct {
	conn network.Conn
	gate *Gate
}

func NewClient(conn network.Conn, gate *Gate) *Client {
	return &Client{
		conn: conn,
		gate: gate,
	}
}

func (p *Client) ReadLoop() {
	for {
		data, err := p.conn.ReadMsg()
		if err != nil {
			fmt.Printf("read message: %v", err)
			break
		}

		fmt.Println(string(data))
		p.WriteMsg("hello")
	}
}

func (p *Client) OnClose() {
	fmt.Println("client close")
	p.gate.Post(func() {
		p.gate.closeEvent(p)
	})
}

func (p *Client) WriteMsg(msg interface{}) {
	str := "hello world"
	data := []byte(str)
	err := p.conn.WriteMsg(data)
	if err != nil {
		fmt.Printf("write message %v error: %v", reflect.TypeOf(msg), err)
	}
}
