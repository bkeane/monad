[private]
default:
    @just --list --unsorted

# build and install monad to ~/.local/bin
install:
    go build -o ~/.local/bin/monad cmd/monad/main.go

# setup docker buildx builder
builder-up:
    docker buildx create --driver=docker-container --name=monad-builder --use

builder-down:
    docker buildx rm monad-builder

[private]
build-echo:
    TAG=$(monad ecr tag --service echo) \
    EPOCH=$(git log -1 --pretty=%ct) \
    docker buildx bake --progress=plain --load

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

    # remove all svg diagrams
    rm .docs/vue/assets/diagrams/*.png

    # generate deployment diagrams
    for file in .docs/mermaid/deployments/*.md; do
        npx github:mermaid-js/mermaid-cli \
        --iconPacks @iconify-json/bitcoin-icons @iconify-json/logos \
        --theme dark \
        --backgroundColor transparent \
        --input $file \
        --cssFile $(dirname $file)/style.css \
        --outputFormat png \
        --output .docs/vue/assets/diagrams/$(basename $file .md).png
    done

    # generate git diagrams
    for file in .docs/mermaid/git/*.md; do
         npx github:mermaid-js/mermaid-cli \
        --iconPacks @iconify-json/bitcoin-icons @iconify-json/logos \
        --theme neutral \
        --backgroundColor transparent \
        --input $file \
        --outputFormat png \
        --output .docs/vue/assets/diagrams/$(basename $file .md).png
    done

    # trim all empty space from the diagrams
    mogrify -trim +repage .docs/vue/assets/diagrams/*.png