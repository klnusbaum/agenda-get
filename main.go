package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/user"
	"path"
	"time"

	"github.com/klnusbaum/agenda-get/errcol"
	"github.com/klnusbaum/agenda-get/orch"
	"github.com/klnusbaum/agenda-get/prog"
	"github.com/klnusbaum/agenda-get/sites"
)

var simpleSites = []sites.SimpleSite{
	sites.Oakland(),
	sites.Bakersfield(),
	sites.SanFrancisco(),
	sites.Pasadena(),
}

var version bool
var outDir string

func init() {
	flag.BoolVar(&version, "version", false, "the version of agenda-get")
	flag.StringVar(&outDir, "out", "", "where agendas should be output")

}

func main() {
	flag.Parse()
	if version {
		printVersion()
		return
	}
	if outDir == "" {
		defaultDir, err := defaultOutDir()
		if err != nil {
			fatalExit(err)
		}
		outDir = defaultDir
	}

	errCol := errcol.Default()
	prog := prog.NewDefault()
	fetcher := sites.NewDefaultFetcher(&http.Client{}, time.Now())
	orch := orch.NewDefault(simpleSites, fetcher, prog, outDir, errCol)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := orch.Run(ctx); err != nil {
		fatalExit(err)
	}
}

func defaultOutDir() (string, error) {
	user, err := user.Current()
	if err != nil {
		return "", fmt.Errorf("can't get current user: %s", err)
	}
	defaultDir := path.Join(user.HomeDir, "agendas")
	if err := os.RemoveAll(defaultDir); err != nil {
		return "", fmt.Errorf("cant clear output directory: %s", err)
	}
	if err := os.MkdirAll(defaultDir, 0755); err != nil {
		return "", fmt.Errorf("can't make agenda directory: %s", err)
	}
	return defaultDir, nil
}

func fatalExit(err error) {
	fmt.Fprintf(os.Stderr, "%s", err)
	os.Exit(1)
}

func printVersion() {
	fmt.Println("TODO - add version logic")
}
