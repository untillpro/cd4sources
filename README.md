# directcd

Continious delivery directly to containers pulling sources from source-repos

# Pulling Single Repo

```sh
./directcd pull --repo https://github.com/untillpro/directcd-test \
  -o directcd-test.exe \
  -t 10 \
  -w .tmp
```

- Repo directcd-test is pulled every 10 seconds
- If changes occur
    - Prevous instance (if any) of `directcd-test.exe` is terminated
    - `go build -o directcd-test.exe` is invoked
    - `directcd-test.exe` is launched
- Working files are located in `.tmp` folder

# Pulling Few Repos

```sh
./directcd pull \
  --repo https://github.com/untillpro/directcd-test \
  --replace https://github.com/untillpro/directcd-test-print=https://github.com/maxim-ge/directcd-test-print \
  -v \
  -o out.exe \
  -t 10 \
  -w .tmp \
  -- --option1 arg1 arg2
```

- Repos specified by `--repo` (main repo) and `--replace` flags are pulled every 10 seconds
- If changes occur 
  - Prevous instance (if any) of `out.exe` is terminated
  - `replace` directive is added to main repo `go.mod`
    - `replace github.com/untillpro/directcd-test-print => ../directcd-test-print`
  - main repo is built and launched
- `-v` means verbose mode
- `--option1 arg1 arg2` are passed to `out.exe`

# Links

Hooks
- https://developer.github.com/v3/repos/hooks/#create-a-hook
- https://developer.github.com/v3/activity/events/types/#pushevent
- https://docs.gitea.io/en-us/webhooks

Stop process
- https://www.ctl.io/developers/blog/post/gracefully-stopping-docker-containers/
  - `docker stop ----time=30 foo`, SIGTERM
  - `docker kill ----signal=SIGQUIT nginx`, if you want to initiate a graceful shutdown of an nginx server, you should send a SIGQUIT
  - `docker kill ----signal=SIGWINCH apache`, Apache uses SIGWINCH to trigger a graceful shutdown
- https://husobee.github.io/golang/ecs/2016/05/19/ecs-graceful-go-shutdown.html

Misc 

- golang url https://play.golang.org/p/6kBtuHvUlQc