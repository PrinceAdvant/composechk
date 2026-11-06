# composechk

A small Go CLI and library that performs lightweight structural checks on Docker Compose YAML files without talking to the Docker daemon. It catches empty files, missing `services`, services without `image` or `build`, duplicate service names, likely unpinned images and localhost references that often break in containers.

## Why

`docker compose config` is the authority when Docker is installed. CI jobs, editors and quick reviews sometimes need a faster, dependency-free pass that fails early on obvious mistakes before a runner pulls images. composechk fills that gap. It is intentionally shallow: it does not replace schema validation or `docker compose config`.

## Install

```
go install github.com/PrinceAdvant/composechk/cmd/composechk@latest
```

## Usage

```
composechk docker-compose.yml
composechk compose.yaml compose.prod.yaml
```

Exit code 1 if any finding has level `error`. Warnings (unpinned tags, localhost) do not fail the process by themselves unless you wrap the tool.

As a library:

```go
r, err := composechk.CheckFile("compose.yaml")
for _, f := range r.Findings {
    fmt.Println(f.Level, f.Message)
}
```

## Good habits it encourages

- Every service has an `image` or a `build` section.
- Image tags are pinned in production Compose files.
- Database URLs use Compose service DNS names, not localhost.
- Run `docker compose config` in CI as the deeper check.

## Where this fits

Structural checks belong next to deploy docs for apps you host yourself. For a fuller production checklist around Compose readiness, see [Docker Compose healthchecks](https://peon.sh/blogs/docker-compose-healthchecks). Maintained by people who work on [Peon](https://peon.sh/), an open-source self-hosted deployment platform.

MIT licensed.
