locals {
    private_subnet_ids = [
        "subnet-0136c58f13b5f8bf9",
        "subnet-00768158825c1f939"
    ]

    security_group_ids = [
        "sg-0102ad4ccceac2613"
    ]
}

data "aws_caller_identity" "current" {}

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
                issuer = "https://kaixo.auth0.com/",
                audience = ["https://kaixo.io"]
            }
        }
    }
}

module "boundary" {
    source = "../modules/boundary"
}

module "hub" {
    source = "github.com/bkeane/monad-action//modules/hub?ref=main"
    # source = "../../../../monad-action/modules/hub"
    depends_on = [aws_iam_openid_connect_provider.github]
    origin = "https://github.com/bkeane/monad.git"
    spoke_account_ids = ["831926600600"]
    boundary_policy_document = module.boundary
    
    services = {
        "e2e/echo" = {
            monad_deploy_args = "--api kaixo --rule file://rule.json --policy file://policy.json"
        }
    }


}

module "spoke" {
    source = "github.com/bkeane/monad-action//modules/spoke?ref=main"
    # source = "../../../../monad-action/modules/spoke"
    depends_on = [aws_iam_openid_connect_provider.github ]
    origin = "https://github.com/bkeane/monad.git"
    api_gateway_ids = toset([module.api_gateway.api_id])
    boundary_policy_document = module.boundary
}

resource "local_file" "deploy" {
    content  = module.hub.deploy
    filename = "../../../.github/workflows/deploy.yml"
}

resource "local_file" "destroy" {
    content  = module.hub.destroy
    filename = "../../../.github/workflows/destroy.yml"
}


