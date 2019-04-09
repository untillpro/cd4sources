# directcd

Continious delivery directly to containers pulling sources from source-repos

# Direct Pulling

In this scenario `directcd` pulls source-code by itself, builds and runs delivered application.

Example: `directdc pull --repo <main-repo> --replace <repo2>=<repo3> -o <binary-name>`

directcd:

- Pull sources from repo1 and repo3
- Every time changes come with pull command:
  - `go.mod` is modified using `replace repo2=>../<repo3-folder-name>`
  - `go build -o <binary-name>` is executed
  - `go.mod` is reverted
  - Previous instance of `<binary-name>` is terminated
  - New instance of `<binary-name>` is launched

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