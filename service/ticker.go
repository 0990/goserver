package service

import "time"

type workTicker struct {
	done     chan struct{}
	worker   Worker
	duration time.Duration
	f        func()
}

func newWorkTicker(worker Worker, d time.Duration, f func()) *workTicker {
	return &workTicker{
		done:     make(chan struct{}, 1),
		worker:   worker,
		duration: d,
		f:        f,
	}
}

func (p *workTicker) run() {
	go func() {
		ticker := time.NewTicker(p.duration)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				p.worker.Post(p.f)
			case <-p.done:
				return
			}
		}
	}()
}

func (p *workTicker) Close() error {
	close(p.done)
	return nil
}
