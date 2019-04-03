package main

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"os/signal"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/spf13/cobra"
	gc "github.com/untillpro/gochips"
)

var binaryName string
var workingDir string
var timeoutSec int32

func pullerCycle(ctx context.Context, wg *sync.WaitGroup, repos map[string]string) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered in pullerCycle", r)
		}
	}()

	os.MkdirAll(workingDir, 655)

	// *************************************************
	gc.Verbose("pullerCycle", "Checking if repos should be cloned")

	for repo, fork := range repos {
		var curRepo string
		if len(fork) > 0 {
			curRepo = fork
		} else {
			curRepo = repo
		}

		u, err := url.Parse(curRepo)
		gc.PanicIfError(err)
		gc.Verbose("pullerCycle", "repo url", u.Path)
		urlPart := strings.Split(u.Path, "/")

		repoSubdir := urlPart[len(urlPart)-1]

		gc.Verbose("pullerCycle", "repoSubdir", repoSubdir)
	}

}

func puller(ctx context.Context, wg *sync.WaitGroup, repos map[string]string, timeout time.Duration) {
	defer wg.Done()

	// *************************************************
	gc.Doing("Pulling repos")
	gc.Info(repos)
	gc.Info("Timeout", timeout)

F:
	for {
		pullerCycle(ctx, wg, repos)
		select {
		case <-time.After(timeout):
		case <-ctx.Done():
			gc.Verbose("puller", "Done")
			break F
		}
	}

	gc.Info("Puller ended")
}

func runPull(cmd *cobra.Command, args []string) {

	// *************************************************
	gc.Doing("Calculating puller parameters")

	re := regexp.MustCompile(`([^=]*)(?:=(.*))?`)
	repos := make(map[string]string)
	for _, arg := range args {
		matches := re.FindStringSubmatch(arg)
		gc.Verbose("arg", arg)
		gc.Verbose("matches", matches)
		if len(matches) > 0 {
			repos[matches[1]] = matches[2]
		}
	}
	gc.Verbose("repos", repos)

	ctx, cancel := context.WithCancel(context.Background())

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)

	var wg sync.WaitGroup

	// *************************************************
	gc.Doing("Starting puller")
	wg.Add(1)
	go puller(ctx, &wg, repos, time.Duration(timeoutSec)*time.Second)

	go func() {
		signal := <-signals
		fmt.Println("Got signal:", signal)
		cancel()
	}()

	// *************************************************
	gc.Doing("Waiting puller")
	wg.Wait()

	gc.Info("Finished")
}
