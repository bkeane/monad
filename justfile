export AWS_PROFILE := "prod.kaixo.io"

branch := `git rev-parse --abbrev-ref HEAD`
sha := `git rev-parse HEAD`

monad := "go run cmd/monad/main.go"
publish := "docker compose -f - build --push"
scaffolds := "go python node ruby"

[private]
default:
    @just --list

# install monad to ~/.local/bin
install:
    go build -o ~/.local/bin/monad cmd/monad/main.go

# login to registry
login:
    aws ecr get-login-password --region us-west-2 | docker login --username AWS --password-stdin 677771948337.dkr.ecr.us-west-2.amazonaws.com

# init scaffolds
init:
    #! /usr/bin/env bash
    for scaffold in {{scaffolds}}; do
        {{monad}} init $scaffold e2e/stage/$scaffold
    done

# publish scaffolds
publish:
    #! /usr/bin/env bash
    for scaffold in {{scaffolds}}; do
        {{monad}} --chdir e2e/stage/$scaffold compose | {{publish}}
    done

# deploy scaffolds
deploy:
    #! /usr/bin/env bash
    for scaffold in {{scaffolds}}; do
        {{monad}} --chdir e2e/stage/$scaffold deploy --api kaixo --auth aws_iam
    done

# test scaffolds
test:
    shellspec --chdir e2e

# destroy scaffolds
destroy:
    #! /usr/bin/env bash
    for scaffold in {{scaffolds}}; do
        {{monad}} --chdir e2e/stage/$scaffold destroy
    done

# clean up scaffolds
clean:
    rm -rf e2e/stage

# Apply terraform
terraform: 
    AWS_PROFILE=prod.kaixo.io tofu -chdir=e2e/terraform/prod init
    AWS_PROFILE=prod.kaixo.io tofu -chdir=e2e/terraform/prod apply

    AWS_PROFILE=dev.kaixo.io tofu -chdir=e2e/terraform/dev init
    AWS_PROFILE=dev.kaixo.io tofu -chdir=e2e/terraform/dev apply



