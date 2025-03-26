locals {
  api_name = "kaixo"
  image = "bkeane/monad/echo"
  common_config = [
    "--image", local.image,
    "--disk", "1024",
    "--memory", "256",
    "--timeout", "10",
    "--api", local.api_name,
    "--policy", "file://e2e/echo/policy.json.tmpl",
    "--rule", "file://e2e/echo/rule.json.tmpl",
    "--env", "file://e2e/echo/.env.tmpl"
  ]
  vpc_config = [
    "--sg-ids", "basic",
    "--sn-ids", "private-a,private-b"
  ]
}

data "aws_caller_identity" "current" {}
data "aws_region" "current" {}

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

module "extended" {
  source          = "../modules/extended"
  account_id      = data.aws_caller_identity.current.account_id
  region          = data.aws_region.current.name
  api_gateway_ids = [module.api_gateway.api_id]
}

module "hub" {
  # source = "github.com/bkeane/monad-action//modules/hub?ref=main"
  source                   = "../../../../monad-action/modules/hub"
  depends_on               = [aws_iam_openid_connect_provider.github]
  origin                   = "https://github.com/bkeane/monad.git"
  boundary_policy_document = module.boundary

  spoke_accounts        = [
    {
      id = data.aws_caller_identity.current.account_id
      name = "prod"
      branches = ["main"]
    },
    {
      id = "831926600600"
      name = "dev"
      branches = ["*"]
    }
  ]

  images = [
    local.image
  ]

  services = [
    {
      name = "echo"
      deploy_cmd = concat(["deploy"], local.common_config)
    },
    {
      name = "echo-oauth"
      deploy_cmd = concat(["deploy", "--auth", "auth0"], local.common_config)
    },
    {
      name = "echo-vpc"
      deploy_cmd = concat(["deploy"], local.common_config, local.vpc_config)
    }
  ]
}

module "spoke" {
  # source = "github.com/bkeane/monad-action//modules/spoke?ref=main"
  source                   = "../../../../monad-action/modules/spoke"
  depends_on               = [aws_iam_openid_connect_provider.github]
  origin                   = "https://github.com/bkeane/monad.git"
  api_gateway_ids          = toset([module.api_gateway.api_id])
  boundary_policy_document = module.boundary
  extended_policy_document = module.extended
}

resource "local_file" "up" {
  content  = module.hub.up
  filename = "../../../.github/workflows/up.yml"
}

resource "local_file" "down" {
  content  = module.hub.down
  filename = "../../../.github/workflows/down.yml"
}

resource "local_file" "untag" {
  content  = module.hub.untag
  filename = "../../../.github/workflows/untag.yml"
}

resource "local_file" "build" {
  content  = module.hub.build
  filename = "../../../.github/actions/monad-build-setup/action.yml"
}

