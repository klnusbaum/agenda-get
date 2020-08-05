package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/user"
	"path"
	"sync"
	"time"

	"github.com/klnusbaum/agenda-get/errcol"
	"github.com/klnusbaum/agenda-get/progress"
	"github.com/klnusbaum/agenda-get/sites"
)

var simpleSites = []sites.SimpleSite{
	sites.Oakland(),
	sites.Bakersfield(),
	sites.Fresno(),
	sites.SanFrancisco(),
	sites.Pasadena(),
}

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
	numSites := len(simpleSites)
	wg := sync.WaitGroup{}
	wg.Add(numSites)
	collector := errcol.Default()
	prog := progress.New(numSites)
	fetcher := sites.NewDefaultFetcher(&http.Client{}, time.Now())

	for _, s := range simpleSites {
		go func(ctx context.Context, s sites.SimpleSite) {
			defer wg.Done()
			defer prog.Increment()
			agenda, err := fetcher.Simple(ctx, s)
			if err != nil {
				collector.Err(err)
				return
			}
			if err := saveAgenda(agenda, outDir); err != nil {
				collector.Err(err)
				return
			}
		}(ctx, s)
	}

	wg.Wait()
	prog.Stop()
	collector.ForEach(func(err error) {
		handlErr(err, outDir)
	})
}

func handlErr(err error, outDir string) {
	var fErr *sites.FinderError
	if errors.As(err, &fErr) {
		reportFinderError(fErr, outDir)
		return
	}
	fmt.Printf("%s\n", err)
}

func reportFinderError(fErr *sites.FinderError, outDir string) {
	filename := path.Join(outDir, fErr.Filename())
	if err := ioutil.WriteFile(filename, fErr.HTML(), 0644); err != nil {
		fatalExit(fmt.Sprintf("report finder error: %s", err))
	}
	fmt.Printf("Critical Error. Please a github issue and attach the %s file.\n", filename)
}

func fatalExit(msg string) {
	fmt.Fprintf(os.Stderr, msg)
	os.Exit(1)
}

func saveAgenda(agenda sites.Agenda, outDir string) error {
	defer agenda.Content.Close()
	outFile, err := os.Create(path.Join(outDir, agenda.Name))
	if err != nil {
		return fmt.Errorf("create output: %s", err)
	}
	if _, err := io.Copy(outFile, agenda.Content); err != nil {
		return fmt.Errorf("write output: %s", err)
	}
	return nil
}
