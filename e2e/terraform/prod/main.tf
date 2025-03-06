locals {
    private_subnet_ids = [
        "subnet-0136c58f13b5f8bf9",
        "subnet-00768158825c1f939"
    ]

    security_group_ids = [
        "sg-0102ad4ccceac2613"
    ]
}

module "api_gateway" {
    source = "terraform-aws-modules/apigateway-v2/aws"

    name          = "kaixo"
    description   = "gateway for mounting functions to prod.kaixo.io"
    protocol_type = "HTTP"

    hosted_zone_name = "prod.kaixo.io"
    domain_name      = "prod.kaixo.io"

    authorizers = {
        "auth0" = {
            name = "auth0"
            authorizer_type = "JWT"
            identity_sources = ["$request.header.Authorization"]
            jwt_configuration = {
                issuer = "https://kaixo.us.auth0.com/",
                audience = ["https://kaixo.io"]
            }
        }
    }
}

module "monad" {
    # source = "github.com/bkeane/monad-action//module?ref=main"
    source = "../../../../monad-action/module"
    origin = "https://github.com/bkeane/monad.git"
    ecr_hub_account_id = "677771948337"
    ecr_spoke_account_ids = ["831926600600"]
    create_oidc_provider = false

    apigatewayv2_ids = toset([module.api_gateway.api_id])

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