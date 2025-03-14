[private]
default:
    @just --list --unsorted

# install monad to ~/.local/bin
install:
    go build -o ~/.local/bin/monad cmd/monad/main.go

# run e2e/spec
test:
    shellspec --chdir e2e

# apply e2e/terraform
terraform: 
    AWS_PROFILE=prod.kaixo.io tofu -chdir=e2e/terraform/prod init
    AWS_PROFILE=prod.kaixo.io tofu -chdir=e2e/terraform/prod apply
    AWS_PROFILE=dev.kaixo.io tofu -chdir=e2e/terraform/dev init
    AWS_PROFILE=dev.kaixo.io tofu -chdir=e2e/terraform/dev apply

