locals {
  api_name = "kaixo"

  security_group_names = [
    "basic"
  ]

  subnet_names = [
    "private-a",
    "private-b"
  ]

  service_common = {
    "MONAD_CHDIR"   = "e2e/echo"
    "MONAD_IMAGE"   = "bkeane/monad/echo"
    "MONAD_API"     = local.api_name
    "MONAD_POLICY"  = "file://policy.json.tmpl"
    "MONAD_RULE"    = "file://rule.json.tmpl"
    "MONAD_ENV"     = "file://.env.tmpl"
  }

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

  name          = local.api_name
  description   = "gateway for mounting functions to prod.kaixo.io"
  protocol_type = "HTTP"

  hosted_zone_name = "prod.kaixo.io"
  domain_name      = "prod.kaixo.io"

  authorizers = {
    "auth0" = {
      name             = "auth0"
      authorizer_type  = "JWT"
      identity_sources = ["$request.header.Authorization"]
      jwt_configuration = {
        issuer   = "https://kaixo.auth0.com/",
        audience = ["https://kaixo.io"]
      }
    }
  }
}

module "boundary" {
  source = "../modules/boundary"
}

module "hub" {
  # source = "github.com/bkeane/monad-action//modules/hub?ref=main"
  source                   = "../../../../monad-action/modules/hub"
  depends_on               = [aws_iam_openid_connect_provider.github]
  origin                   = "https://github.com/bkeane/monad.git"
  spoke_account_ids        = [data.aws_caller_identity.current.account_id, "831926600600"]
  boundary_policy_document = module.boundary

  services = [
    merge(local.service_common, {
      "MONAD_SERVICE" = "echo"
    }),
    merge(local.service_common, {
      "MONAD_SERVICE" = "echo-oauth"
      "MONAD_AUTH"    = "auth0"
    }),
    merge(local.service_common, {
      "MONAD_SERVICE"         = "echo-vpc"
      "MONAD_SECURITY_GROUPS" = join(",", local.security_group_names)
      "MONAD_SUBNETS"         = join(",", local.subnet_names)
    })
  ]
}

module "spoke" {
  # source = "github.com/bkeane/monad-action//modules/spoke?ref=main"
  source                   = "../../../../monad-action/modules/spoke"
  depends_on               = [aws_iam_openid_connect_provider.github]
  origin                   = "https://github.com/bkeane/monad.git"
  api_gateway_ids          = toset([module.api_gateway.api_id])
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


