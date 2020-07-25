package progress

import (
	"fmt"
	"sync"
	"time"
)

type Progress struct {
	sync.RWMutex
	total    int
	complete int
	stopCh   chan struct{}
	done     sync.WaitGroup
}

func Start(total int) *Progress {
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
			p.writeProgress()
		case <-p.stopCh:
			return
		}
	}
}

func (p *Progress) Stop() {
	p.stopCh <- struct{}{}
	p.done.Wait()
	p.writeProgress()
	fmt.Println()
}

func (p *Progress) writeProgress() {
	p.RLock()
	defer p.RUnlock()
	percent := (float32(p.complete) / float32(p.total)) * 100.0
	fmt.Printf("\r%.f%% complete", percent)
}

func (p *Progress) Increment() {
	p.Lock()
	defer p.Unlock()
	p.complete++
}
