module "monad" {
    source = "github.com/bkeane/monad-action//module?ref=main"
    origin = "https://github.com/bkeane/monad.git"
    ecr_hub_account_id = "677771948337"
    ecr_spoke_account_ids = ["831926600600"]
    create_oidc_provider = false

    services = {
        "e2e/echo" = {}
    }
}

resource "local_file" "deploy" {
    content  = module.monad.deploy
    filename = "../../../.github/workflows/deploy.yml"
}

resource "local_file" "destroy" {
    content  = module.monad.destroy
    filename = "../../../.github/workflows/destroy.yml"
}