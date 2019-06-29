package network

type Clienter interface {
	ReadLoop()
	OnClose()
}
