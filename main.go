package main

import (
	"context"
	"fmt"
	"sync"
	"time"
)

var sites = []site{
	simpleSite{"blah.com", 1},
	simpleSite{"foo.com", 4},
	simpleSite{"bar.com", 6},
	simpleSite{"baz.com", 10},
}

type site interface {
	get(context.Context)
}

type simpleSite struct {
	url      string
	duration time.Duration
}

func (s simpleSite) get(ctx context.Context) {
	time.Sleep(s.duration * time.Second)
}

type progress struct {
	sync.RWMutex
	total    int
	complete int
	stopCh   chan struct{}
	done     sync.WaitGroup
}

func start(total int) *progress {
	p := &progress{
		total:  total,
		stopCh: make(chan struct{}, 1),
	}
	p.done.Add(1)
	go p.run()
	return p
}

func (p *progress) run() {
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

func (p *progress) stop() {
	p.stopCh <- struct{}{}
	p.done.Wait()
	p.writeProgress()
	fmt.Println()
}

func (p *progress) writeProgress() {
	p.RLock()
	defer p.RUnlock()
	percent := (float32(p.complete) / float32(p.total)) * 100.0
	fmt.Printf("\r%.f%% complete", percent)
}

func (p *progress) increment() {
	p.Lock()
	defer p.Unlock()
	p.complete++
}

func main() {
	numSites := len(sites)

	wg := sync.WaitGroup{}
	wg.Add(numSites)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	prog := start(numSites)
	defer prog.stop()

	for _, s := range sites {
		go func(ctx context.Context, s site) {
			defer wg.Done()
			defer prog.increment()
			s.get(ctx)
		}(ctx, s)
	}

	wg.Wait()
}
