package progress

import (
	"fmt"
	"sync"
	"time"
)

const _barWidth = 20

type Progress struct {
	sync.RWMutex
	total    int
	complete int
	stopCh   chan struct{}
	done     sync.WaitGroup
}

func New(total int) *Progress {
	p := &Progress{
		total:  total,
		stopCh: make(chan struct{}, 1),
	}
	p.done.Add(1)
	go p.run()
	return p
}

func (p *Progress) run() {
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

func (p *Progress) Stop() {
	p.stopCh <- struct{}{}
	p.done.Wait()
}

func (p *Progress) draw() {
	p.RLock()
	defer p.RUnlock()
	percent := (float32(p.complete) / float32(p.total)) * 100.0
	bar := p.drawBar()
	fmt.Printf("\r|%3.f%% complete %s|", percent, bar)
}

func (p *Progress) drawBar() string {
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

func (p *Progress) Increment() {
	p.Lock()
	defer p.Unlock()
	p.complete++
}
