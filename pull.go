package main

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/url"
	"os"
	"os/exec"
	"os/signal"
	"path"
	"path/filepath"
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
var mappedRepos []string

type puller struct {
	ctx              context.Context
	repos            []string
	substs           map[string]string
	timeout          time.Duration
	lastCommitHashes map[string]string
	cmd              *exec.Cmd
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

// <workingDir>/<repoFolder>
// <repoPath               >
// repoPath = workingDir + '/' + repoFolder
func (p *puller) getRepoFolders(repoURL string) (repoPath string, repoFolder string) {
	u, err := url.Parse(repoURL)
	gc.PanicIfError(err)
	urlParts := strings.Split(u.Path, "/")
	repoFolder = urlParts[len(urlParts)-1]
	repoPath = path.Join(workingDir, repoFolder)
	return
}

func (p *puller) stopCmd() {
	defer func() { p.cmd = nil }()
	if nil != p.cmd {
		gc.Doing("stopCmd: Terminating  previous process")
		err := p.cmd.Process.Kill()
		if nil != err {
			gc.Info("stopCmd: Error occured:", err)
		}
		gc.Info("Done")
	}
}

func (p *puller) iteration() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("iteration: Recovered: ", r)
		}
	}()

	// *************************************************
	gc.Verbose("iteration:", "Checking if repos should be cloned")

	atLeastOneRepoChanged := false

	for _, curRepoURL := range p.repos {

		repoPath, repoFolder := p.getRepoFolders(curRepoURL)

		gc.Verbose("iteration:", "repoPath, repoFolder=", repoPath, repoFolder)

		os.MkdirAll(workingDir, 0755)

		if _, err := os.Stat(repoPath); os.IsNotExist(err) {
			gc.Info("iteration:", "Repo folder does not exist, will be cloned", repoPath, curRepoURL)
			err := new(gc.PipedExec).
				Command("git", "clone", curRepoURL).
				WorkingDir(workingDir).
				Run(os.Stdin, os.Stderr)
			gc.PanicIfError(err)
		} else {
			gc.Verbose("iteration:", "Repo dir exists, will be pulled", repoPath, curRepoURL)
			stdouts, stderrs, err := new(gc.PipedExec).
				Command("git", "pull", curRepoURL).
				WorkingDir(repoPath).
				RunToStrings()
			if nil != err {
				gc.Info(stdouts, stderrs)
			}
			gc.PanicIfError(err)
		}

		newHash := getLastCommitHash(repoPath)
		oldHash := p.lastCommitHashes[repoPath]
		if oldHash == newHash {
			gc.Verbose("*** Nothing changed")
		} else {
			gc.Info("iteration:", "Commit hash changed", oldHash, newHash)
			atLeastOneRepoChanged = true
			p.lastCommitHashes[repoPath] = newHash
		}
	} // for repors

	if atLeastOneRepoChanged {

		p.stopCmd()

		gc.Info("iteration:", "Main repo will be rebuilt")

		repoPath, _ := p.getRepoFolders(p.repos[0])

		gc.Doing("go get -u")
		err := new(gc.PipedExec).
			Command("go", "get", "-u").
			WorkingDir(repoPath).
			Run(verboseWriters())
		gc.PanicIfError(err)

		gc.Doing("go build")
		err = new(gc.PipedExec).
			Command("go", "build", "-o", binaryName).
			WorkingDir(repoPath).
			Run(verboseWriters())
		gc.PanicIfError(err)

		gc.Info("iteration:", "Done")

		fileToExec, err := filepath.Abs(path.Join(repoPath, binaryName))
		gc.PanicIfError(err)
		gc.Doing("Running " + fileToExec)

		pe := new(gc.PipedExec)
		err = pe.Command(fileToExec).
			WorkingDir(repoPath).
			Start(os.Stdout, os.Stderr)
		gc.PanicIfError(err)
		p.cmd = pe.GetCmd(0)
		gc.Info("iteration:", "Process started!")
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
			p.stopCmd()
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
	for _, arg := range mappedRepos {
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
