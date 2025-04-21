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
    docker run -t --env-file <(env | grep -E '(MONAD|AWS)') -v $(pwd):/src ghcr.io/bkeane/spec:latest --chdir e2e

# apply e2e/terraform
terraform: 
    AWS_PROFILE=prod.kaixo.io tofu -chdir=e2e/terraform/prod init
    AWS_PROFILE=prod.kaixo.io tofu -chdir=e2e/terraform/prod apply
    AWS_PROFILE=dev.kaixo.io tofu -chdir=e2e/terraform/dev init
    AWS_PROFILE=dev.kaixo.io tofu -chdir=e2e/terraform/dev apply

# open docs in browser for development
docs: 
    cd docs/vue && npm run dev

# build diagrams
diagrams:
    #! /usr/bin/env bash
    
    # We are using the main branch of mermaid-cli for icon support.
    # In the future, when this is in a release, we should lock to a specific version.
    
    rm docs/vue/assets/diagrams/*.svg

    for file in docs/mermaid/deployments/*.md; do
        npx github:mermaid-js/mermaid-cli \
        --iconPacks @iconify-json/bitcoin-icons @iconify-json/logos \
        --theme dark \
        --backgroundColor transparent \
        --cssFile $(dirname $file)/style.css \
        --input $file \
        --output docs/vue/assets/diagrams/$(basename $file .md).svg
    done

    for file in docs/mermaid/git/*.md; do
         npx github:mermaid-js/mermaid-cli \
        --iconPacks @iconify-json/bitcoin-icons @iconify-json/logos \
        --theme neutral \
        --backgroundColor transparent \
        --input $file \
        --output docs/vue/assets/diagrams/$(basename $file .md).svg
    done
