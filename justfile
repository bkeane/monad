export AWS_PROFILE := "prod.kaixo.io"

protos := "proto/*.proto"

branch := `git rev-parse --abbrev-ref HEAD`
sha := `git rev-parse HEAD`

monad := "go run cmd/monad/main.go"
publish := "docker compose -f - build --push"
scaffolds := "go python node ruby"

[private]
default:
    @just --list

# generate protobuf
gen:
    protoc --go_out=. {{protos}}

# install monad to ~/.local/bin
install:
    go build -o ~/.local/bin/monad cmd/monad/main.go

# login to registry
login:
    #! /usr/bin/env bash
    aws ecr get-login-password --region us-west-2 \
    | docker login --username AWS --password-stdin 677771948337.dkr.ecr.us-west-2.amazonaws.com

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
        {{monad}} encode --context e2e/stage/$scaffold | {{publish}}
    done

# deploy scaffolds
deploy:
    #! /usr/bin/env bash
    for scaffold in {{scaffolds}}; do
        {{monad}} deploy --context e2e/stage/$scaffold -f api -d spoke
    done

# test scaffolds
test:
    #! /usr/bin/env bash
    cd e2e
    for scaffold in {{scaffolds}}; do
        bundle exec ruby test.rb stage/$scaffold -p personal -f api
    done

# destroy scaffolds
destroy:
    #! /usr/bin/env bash
    for scaffold in {{scaffolds}}; do
        LOG_LEVEL=info {{monad}} destroy --context e2e/stage/$scaffold -d spoke platform
    done

# tail logs for given scaffold
tail scaffold:
    aws logs tail /aws/lambda/monad-{{branch}}-{{scaffold}} --follow

# clean up scaffolds
clean:
    rm -rf e2e/stage




