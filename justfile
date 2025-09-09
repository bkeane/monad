# documentation
mod docs 'docs/justfile'

export GIT_SHA := shell('git rev-parse HEAD')
export GIT_BRANCH := shell('git rev-parse --abbrev-ref HEAD')

# List help
[private]
default:
    @just --list --unsorted

# build and install monad to ~/.local/bin
install:
    go build -o ~/.local/bin/monad cmd/monad/main.go

# build docker images
build:
    docker buildx bake --progress=plain
