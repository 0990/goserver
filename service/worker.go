package service

import (
	"github.com/0990/goserver/util"
	"runtime/debug"
	"time"
)

type Worker interface {
	Post(f func())
	Run()
	Close()
	AfterPost(duration time.Duration, f func()) *time.Timer
	NewTicker(d time.Duration, f func()) *time.Ticker
	Len() int
}

//TODO 这里的实现 chan如果塞满会阻塞进程，可对比参照github.com/davyxu/cellnet EventQueue实现，选方案
type Work struct {
	funChan chan func()
}

func NewWorker() Worker {
	p := new(Work)
	p.funChan = make(chan func(), 10240)
	return p
}

func (p *Work) Post(f func()) {
	p.funChan <- f
}

func (p *Work) Run() {
	go func() {
		for f := range p.funChan {
			util.ProtectedFun(f)
		}
	}()
}

func (p *Work) Close() {
	close(p.funChan)
}

func (p *Work) Len() int {
	return len(p.funChan)
}
func (p *Work) protectedFun(callback func()) {
	//TODO 每个函数都包装了defer，性能怎样？
	defer func() {
		if err := recover(); err != nil {
			debug.PrintStack()
		}
	}()
	callback()
}

func (p *Work) AfterPost(d time.Duration, f func()) *time.Timer {
	return time.AfterFunc(d, func() {
		p.Post(f)
	})
}

func (p *Work) NewTicker(d time.Duration, f func()) *time.Ticker {
	ticker := time.NewTicker(d)
	go func() {
		for range ticker.C {
			p.Post(f)
			//util.ProtectedFun(f)
		}
	}()
	return ticker
}
