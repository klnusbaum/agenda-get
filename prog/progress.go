package prog

import (
	"fmt"
	"sync"
	"time"
)

const _barWidth = 20

type Progress interface {
	Start(int)
	Increment()
	Stop()
}

type DefaultProgress struct {
	sync.RWMutex
	total    int
	complete int
	stopCh   chan struct{}
	done     sync.WaitGroup
}

func NewDefault() *DefaultProgress {
	p := &DefaultProgress{
		stopCh: make(chan struct{}, 1),
	}
	return p
}

func (p *DefaultProgress) Start(total int) {
	p.total = total
	p.done.Add(1)
	go p.run()
}

func (p *DefaultProgress) run() {
	defer p.done.Done()
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			p.draw()
		case <-p.stopCh:
			p.draw()
			fmt.Println()
			return
		}
	}
}

func (p *DefaultProgress) Stop() {
	p.stopCh <- struct{}{}
	p.done.Wait()
}

func (p *DefaultProgress) draw() {
	p.RLock()
	defer p.RUnlock()
	percent := (float32(p.complete) / float32(p.total)) * 100.0
	bar := p.drawBar()
	fmt.Printf("\r|%3.f%% complete %s|", percent, bar)
}

func (p *DefaultProgress) drawBar() string {
	length := int((float32(p.complete) / float32(p.total)) * _barWidth)
	bar := ""
	for i := 0; i < _barWidth; i++ {
		if i < length {
			bar = bar + "="
		} else {
			bar = bar + " "
		}
	}
	return bar
}

func (p *DefaultProgress) Increment() {
	p.Lock()
	defer p.Unlock()
	p.complete++
}
