# pullcds
Continious delivery using pull requests to source-code repositories

# Use Cases

## Configuring Server

- Working directory

## Quering Version

- Client sends a `watch` request for given version
- Request is blocked for a given timeout until new version is available
- If `watch` request returns new version ID client sends a `download` request

Watch Request
- `http://sourcecds/watch?<params>`
  - repo=https://github.com/mycompany/myrepo`
  - repo=https://github.com/mycompany/myrepo`


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