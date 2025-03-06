locals {
    private_subnet_ids = [
        "subnet-0f00afaf6cb110510",
        "subnet-05129d09890492ffd"
    ]

    security_group_ids = [
        "sg-0018ec3d366c44cc1"
    ]
}

module "api_gateway" {
    source = "terraform-aws-modules/apigateway-v2/aws"

    name          = "kaixo"
    description   = "gateway for mounting functions to dev.kaixo.io"
    protocol_type = "HTTP"

    hosted_zone_name = "dev.kaixo.io"
    domain_name      = "dev.kaixo.io"

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
    create_oidc_provider = true

    apigatewayv2_ids = toset([module.api_gateway.api_id])

    services = {
        "e2e/echo" = {}
    }
}
