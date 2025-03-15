[private]
default:
    @just --list --unsorted

# build and install monad to ~/.local/bin
install:
    go build -o ~/.local/bin/monad cmd/monad/main.go

# Build docker images locally
build:
    # docker build -t ghcr.io/bkeane/monad:latest --platform linux/amd64,linux/arm64 .
    docker build -t ghcr.io/bkeane/shellspec:latest --platform linux/amd64,linux/arm64 --target shellspec .

# approximate github actions tests
test:
    #! /usr/bin/env bash
    session=$(aws sts get-session-token --duration-seconds 3600)
    export AWS_ACCESS_KEY_ID=$(echo $session | jq -r .Credentials.AccessKeyId)
    export AWS_SECRET_ACCESS_KEY=$(echo $session | jq -r .Credentials.SecretAccessKey)
    export AWS_SESSION_TOKEN=$(echo $session | jq -r .Credentials.SessionToken)
    unset AWS_PROFILE
    docker run -it --rm --env-file <(env | grep -E '(MONAD|AWS)') -v $(pwd):/src --workdir /src ghcr.io/bkeane/shellspec:latest --chdir e2e

# apply e2e/terraform
terraform: 
    AWS_PROFILE=prod.kaixo.io tofu -chdir=e2e/terraform/prod init
    AWS_PROFILE=prod.kaixo.io tofu -chdir=e2e/terraform/prod apply
    AWS_PROFILE=dev.kaixo.io tofu -chdir=e2e/terraform/dev init
    AWS_PROFILE=dev.kaixo.io tofu -chdir=e2e/terraform/dev apply

