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
	repos            map[string]string
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
			fmt.Println("Recovered in pullerCycle", r)
		}
	}()

	// *************************************************
	gc.Verbose("pullerCycle", "Checking if repos should be cloned")

	for repo, fork := range p.repos {
		var curRepoURL string
		if len(fork) > 0 {
			curRepoURL = fork
		} else {
			curRepoURL = repo
		}

		// <workingDir>/<repoFolderName>-wd/<repoFolderName>
		// <repoWd                        >/<repoFolderName>
		// <repoDir                                        >

		u, err := url.Parse(curRepoURL)
		gc.PanicIfError(err)
		gc.Verbose("pullerCycle", "repo url", u.Path)
		urlParts := strings.Split(u.Path, "/")

		repoFolderName := urlParts[len(urlParts)-1]
		repoWd := path.Join(workingDir, repoFolderName+"-wd")
		repoDir := path.Join(repoWd, repoFolderName)

		gc.Verbose("pullerCycle", "workingDir", workingDir)
		gc.Verbose("pullerCycle", "repoDir", repoDir)

		os.MkdirAll(repoWd, 0755)

		if _, err := os.Stat(repoDir); os.IsNotExist(err) {
			gc.Info("pullerCycle", "Repo dir does not exist, will be cloned", repoDir, curRepoURL)
			err := new(gc.PipedExec).
				Command("git", "clone", curRepoURL).
				WorkingDir(repoWd).
				Run(os.Stdin, os.Stderr)
			gc.PanicIfError(err)
		} else {
			gc.Verbose("pullerCycle", "Repo dir exists, will be pulled", repoDir, curRepoURL)
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

	// *************************************************
	gc.Doing("Pulling repos")
	gc.Info(p.repos)
	gc.Info("Timeout", p.timeout)

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

	p := &puller{ctx: ctx, repos: repos, timeout: time.Duration(timeoutSec) * time.Second, lastCommitHashes: map[string]string{}}

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
