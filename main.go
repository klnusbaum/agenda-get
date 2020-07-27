package main

import (
	"context"
	"fmt"
	"os"
	"os/user"
	"path"
	"sync"

	"github.com/klnusbaum/agenda-get/errcol"
	"github.com/klnusbaum/agenda-get/progress"
	"github.com/klnusbaum/agenda-get/sites"
)

func main() {
	numSites := len(sites.Sites)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	user, err := user.Current()
	if err != nil {
		fmt.Println("can't get current user")
		os.Exit(1)
	}
	outDir := path.Join(user.HomeDir, "agendas")
	if err := os.RemoveAll(outDir); err != nil {
		fmt.Println("can't clear output directory")
		os.Exit(1)
	}
	if err := os.MkdirAll(outDir, 0755); err != nil {
		fmt.Println("can't make agenda directory")
		os.Exit(1)
	}

	wg := sync.WaitGroup{}
	wg.Add(numSites)
	collector := errcol.Default()
	prog := progress.New(numSites)

	for _, s := range sites.Sites {
		go func(ctx context.Context, s sites.Site) {
			defer wg.Done()
			defer prog.Increment()
			collector.Err(s.Get(ctx, outDir))
		}(ctx, s)
	}

	wg.Wait()
	prog.Stop()

	collector.ForEach(func(err error) {
		fmt.Printf("%s\n", err)
	})
}
