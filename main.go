package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/user"
	"path"
	"sync"

	"github.com/klnusbaum/agenda-get/errcol"
	"github.com/klnusbaum/agenda-get/progress"
	"github.com/klnusbaum/agenda-get/sites"
)

func main() {
	user, err := user.Current()
	if err != nil {
		fatalExit(fmt.Sprintf("can't get current user: %s", err))
	}
	outDir := path.Join(user.HomeDir, "agendas")
	if err := os.RemoveAll(outDir); err != nil {
		fatalExit(fmt.Sprintf("cant clear output directory: %s", err))
	}
	if err := os.MkdirAll(outDir, 0755); err != nil {
		fatalExit(fmt.Sprintf("can't make agenda directory: %s", err))
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	numSites := len(sites.Sites)
	wg := sync.WaitGroup{}
	wg.Add(numSites)
	collector := errcol.Default()
	prog := progress.New(numSites)
	client := &http.Client{}

	for _, s := range sites.Sites {
		go func(ctx context.Context, s sites.Site) {
			defer wg.Done()
			defer prog.Increment()
			collector.Err(s.Get(ctx, client, outDir))
		}(ctx, s)
	}

	wg.Wait()
	prog.Stop()
	collector.ForEach(func(err error) {
		fmt.Printf("%s\n", err)
	})
}

func fatalExit(msg string) {
	fmt.Fprintf(os.Stderr, msg)
	os.Exit(1)
}
