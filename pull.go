package main

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/url"
	"os"
	"os/signal"
	"path"
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

type puller struct {
	ctx              context.Context
	repos            []string
	substs           map[string]string
	timeout          time.Duration
	lastCommitHashes map[string]string
}

func verboseWriters() (out io.Writer, err io.Writer) {
	if gc.IsVerbose {
		return os.Stdout, os.Stderr
	}
	return ioutil.Discard, os.Stderr
}

func getLastCommitHash(repoDir string) string {
	stdout, _, err := new(gc.PipedExec).
		Command("git", "log", "-n", "1", `--pretty=format:%H`).
		WorkingDir(repoDir).
		RunToStrings()
	gc.PanicIfError(err)

	return strings.TrimSpace(stdout)
}

func (p *puller) iteration() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered in puller iteration", r)
		}
	}()

	// *************************************************
	gc.Verbose("iteration", "Checking if repos should be cloned")

	for _, curRepoURL := range p.repos {

		// <workingDir>/<repoFolderName>
		// <repoDir                    >

		u, err := url.Parse(curRepoURL)
		gc.PanicIfError(err)
		gc.Verbose("iteration", "repo url", u.Path)
		urlParts := strings.Split(u.Path, "/")

		repoFolderName := urlParts[len(urlParts)-1]
		repoDir := path.Join(workingDir, repoFolderName)

		gc.Verbose("iteration", "workingDir", workingDir)
		gc.Verbose("iteration", "repoDir", repoDir)

		os.MkdirAll(workingDir, 0755)

		if _, err := os.Stat(repoDir); os.IsNotExist(err) {
			gc.Info("iteration", "Repo dir does not exist, will be cloned", repoDir, curRepoURL)
			err := new(gc.PipedExec).
				Command("git", "clone", curRepoURL).
				WorkingDir(workingDir).
				Run(os.Stdin, os.Stderr)
			gc.PanicIfError(err)
		} else {
			gc.Verbose("iteration", "Repo dir exists, will be pulled", repoDir, curRepoURL)
			stdouts, stderrs, err := new(gc.PipedExec).
				Command("git", "pull", curRepoURL).
				WorkingDir(repoDir).
				RunToStrings()
			if nil != err {
				gc.Info(stdouts, stderrs)
			}
			gc.PanicIfError(err)
		}

		newHash := getLastCommitHash(repoDir)
		oldHash := p.lastCommitHashes[repoDir]
		if oldHash == newHash {
			gc.Verbose("*** Nothing changed")
		} else {
			gc.Info("iteration", "Commit hash changed", oldHash, newHash)
			err := new(gc.PipedExec).
				Command("go", "build", "-o", binaryName).
				WorkingDir(repoDir).
				Run(verboseWriters())
			gc.PanicIfError(err)
		}

		p.lastCommitHashes[repoDir] = newHash
	}

}

func cycle(p *puller, wg *sync.WaitGroup) {
	defer wg.Done()

	gc.Info("Puller started")
	gc.Info("repos", p.repos)
	gc.Info("substs", p.substs)
	gc.Info("timeout", p.timeout)

	// *************************************************

F:
	for {
		p.iteration()
		select {
		case <-time.After(p.timeout):
		case <-p.ctx.Done():
			gc.Verbose("puller", "Done")
			break F
		}
	}

	gc.Info("Puller ended")
}

func runCmdPull(cmd *cobra.Command, args []string) {

	// *************************************************
	gc.Doing("Calculating puller parameters")

	re := regexp.MustCompile(`([^=]*)(?:=(.*))?`)
	repos := []string{}
	substs := make(map[string]string)
	for _, arg := range args {
		matches := re.FindStringSubmatch(arg)
		gc.Verbose("arg", arg)
		gc.Verbose("matches", matches)
		if len(matches[2]) > 0 {
			substs[matches[1]] = matches[2]
			repos = append(repos, matches[2])
		} else {
			repos = append(repos, matches[1])
		}
	}

	ctx, cancel := context.WithCancel(context.Background())

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)

	var wg sync.WaitGroup

	// *************************************************
	gc.Doing("Starting puller")
	wg.Add(1)

	p := &puller{ctx: ctx, repos: repos, substs: substs, timeout: time.Duration(timeoutSec) * time.Second, lastCommitHashes: map[string]string{}}

	go cycle(p, &wg)

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
