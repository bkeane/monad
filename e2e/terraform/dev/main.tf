locals {
    private_subnet_ids = [
        "subnet-0f00afaf6cb110510",
        "subnet-05129d09890492ffd"
    ]

    security_group_ids = [
        "sg-0018ec3d366c44cc1"
    ]
}

resource "aws_iam_openid_connect_provider" "github" {
  url            = "https://token.actions.githubusercontent.com"
  client_id_list = ["sts.amazonaws.com"]

  thumbprint_list = [
    "6938fd4d98bab03faadb97b34396831e3780aea1",
    "1c58a3a8518e8759bf075b76b750d4f2df264fcd"
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
                issuer = "https://kaixo.auth0.com/",
                audience = ["https://kaixo.io"]
            }
        }
    }
}

module "boundary" {
    source = "../modules/boundary"
}

module "spoke" {
    # source = "github.com/bkeane/monad-action//modules/spoke?ref=main"
    source = "../../../../monad-action/modules/spoke"
    depends_on = [aws_iam_openid_connect_provider.github]
    origin = "https://github.com/bkeane/monad.git"
    api_gateway_ids = toset([module.api_gateway.api_id])
    boundary_policy_document = module.boundary
}
