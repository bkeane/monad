[private]
default:
    @just --list --unsorted

# build and install monad to ~/.local/bin
install:
    go build -o ~/.local/bin/monad cmd/monad/main.go

# setup docker buildx builder
builder:
    docker buildx create --driver=docker-container --name=monad-builder --driver-opt default-load=true --use

[private]
build-echo:
    #! /usr/bin/env bash
    cd e2e/echo
    docker buildx build -t $(monad ecr tag) \
    --platform linux/arm64,linux/amd64 \
    --cache-to type=s3,region=us-west-2,bucket=kaixo-buildx-cache,name=echo,mode=max \
    --cache-from type=s3,region=us-west-2,bucket=kaixo-buildx-cache,name=echo \
    .

[private]
build-monad:
    #! /usr/bin/env bash
    docker buildx build -t ghcr.io/bkeane/monad:latest \
    --cache-to type=s3,region=us-west-2,bucket=kaixo-buildx-cache,name=monad,mode=max \
    --cache-from type=s3,region=us-west-2,bucket=kaixo-buildx-cache,name=monad \
    --platform linux/amd64,linux/arm64 .

# Build docker images locally
build: build-monad build-echo

# apply e2e/terraform
terraform: 
    AWS_PROFILE=prod.kaixo.io tofu -chdir=e2e/terraform/prod init
    AWS_PROFILE=prod.kaixo.io tofu -chdir=e2e/terraform/prod apply
    AWS_PROFILE=dev.kaixo.io tofu -chdir=e2e/terraform/dev init
    AWS_PROFILE=dev.kaixo.io tofu -chdir=e2e/terraform/dev apply

# open docs in browser for development
docs: 
    cd .docs/vue && npm run dev

# build docs
build-docs:
    cd .docs/vue && npm run build

# build diagrams
diagrams:
    #! /usr/bin/env bash
    
    # We are using the main branch of mermaid-cli for icon support.
    # In the future, when this is in a release, we should lock to a specific version.
    
    rm .docs/vue/assets/diagrams/*.svg

    for file in .docs/mermaid/deployments/*.md; do
        npx github:mermaid-js/mermaid-cli \
        --iconPacks @iconify-json/bitcoin-icons @iconify-json/logos \
        --theme dark \
        --backgroundColor transparent \
        --cssFile $(dirname $file)/style.css \
        --input $file \
        --output .docs/vue/assets/diagrams/$(basename $file .md).svg
    done

    for file in .docs/mermaid/git/*.md; do
         npx github:mermaid-js/mermaid-cli \
        --iconPacks @iconify-json/bitcoin-icons @iconify-json/logos \
        --theme neutral \
        --backgroundColor transparent \
        --input $file \
        --output .docs/vue/assets/diagrams/$(basename $file .md).svg
    done

