package service

import "runtime/debug"

type Worker interface {
	Post(f func())
	Run()
	Close()
}

//TODO 这里的实现 chan如果塞满会阻塞进程，可对比参照github.com/davyxu/cellnet EventQueue实现，选方案
type Work struct {
	funChan chan func()
}

func NewWorker() Worker {
	p := new(Work)
	p.funChan = make(chan func(), 1024)
	return p
}

func (p *Work) Post(f func()) {
	p.funChan <- f
}

func (p *Work) Run() {
	go func() {
		for f := range p.funChan {
			p.protectedFun(f)
		}
	}()
}

func (p *Work) Close() {
	close(p.funChan)
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
