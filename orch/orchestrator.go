package orch

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"sync"

	"github.com/klnusbaum/agenda-get/errcol"
	"github.com/klnusbaum/agenda-get/prog"
	"github.com/klnusbaum/agenda-get/sites"
)

type Orchestrator interface {
	Run(ctx context.Context) error
}

type DefaultOrchestrator struct {
	simpleSites []sites.SimpleSite
	fetcher     sites.Fetcher
	progress    prog.Progress
	outDir      string
	errCol      errcol.Collector
}

func NewDefault(
	simpleSites []sites.SimpleSite,
	fetcher sites.Fetcher,
	progress prog.Progress,
	outDir string,
	errCol errcol.Collector) DefaultOrchestrator {
	return DefaultOrchestrator{
		simpleSites: simpleSites,
		fetcher:     fetcher,
		progress:    progress,
		outDir:      outDir,
		errCol:      errCol,
	}
}

func (o DefaultOrchestrator) Run(ctx context.Context) error {
	numSites := len(o.simpleSites)
	wg := sync.WaitGroup{}
	wg.Add(numSites)
	o.progress.Start(numSites)

	for _, s := range o.simpleSites {
		go func(ctx context.Context, s sites.SimpleSite) {
			defer wg.Done()
			defer o.progress.Increment()
			agenda, err := o.fetcher.Simple(ctx, s)
			if err != nil {
				o.errCol.Err(err)
				return
			}
			if err := o.saveAgenda(agenda); err != nil {
				o.errCol.Err(err)
				return
			}
		}(ctx, s)
	}

	wg.Wait()
	o.progress.Stop()
	return o.errCol.ForEach(o.handlErr)
}

func (o DefaultOrchestrator) saveAgenda(agenda sites.Agenda) error {
	defer agenda.Content.Close()
	outFile, err := os.Create(path.Join(o.outDir, agenda.Name))
	if err != nil {
		return fmt.Errorf("create output: %s", err)
	}
	if _, err := io.Copy(outFile, agenda.Content); err != nil {
		return fmt.Errorf("write output: %s", err)
	}
	return nil
}

func (o DefaultOrchestrator) handlErr(err error) error {
	var fErr *sites.FinderError
	if errors.As(err, &fErr) {
		return o.reportFinderError(fErr)
	}
	fmt.Printf("%s\n", err)
	return nil
}

func (o DefaultOrchestrator) reportFinderError(fErr *sites.FinderError) error {
	filename := path.Join(o.outDir, fErr.Filename())
	if err := ioutil.WriteFile(filename, fErr.HTML(), 0644); err != nil {
		return fmt.Errorf("report finder error: %s", err)
	}
	fmt.Printf("Critical Error. Please a github issue and attach the %s file.\n", filename)
	return nil
}
