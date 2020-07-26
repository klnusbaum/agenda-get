package main

import (
	"context"
	"sync"
	"time"

	"github.com/klnusbaum/agenda-get/progress"
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

func main() {
	numSites := len(sites)

	wg := sync.WaitGroup{}
	wg.Add(numSites)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	prog := progress.New(numSites)
	defer prog.Stop()

	for _, s := range sites {
		go func(ctx context.Context, s site) {
			defer wg.Done()
			defer prog.Increment()
			s.get(ctx)
		}(ctx, s)
	}

	wg.Wait()
}
