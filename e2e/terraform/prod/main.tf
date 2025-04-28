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
    "--vpc-sg", "basic",
    "--vpc-sn", "private-a private-b"
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

module "topology" {
  source = "../../../../monad-action/modules/topology"
  origin = "https://github.com/bkeane/monad.git"

  enable_boundary_policy = false
  
  integration_account_name = "prod"
  integration_account_id = "677771948337"
  integration_account_ecr_region = "us-west-2"
  
  integration_account_ecr_paths = [
    "bkeane/monad/echo"
  ]

  deployment_accounts = {
    "prod" = "677771948337"
    "dev" = "831926600600"
  }
}

module "integration" {
  # source = "github.com/bkeane/monad-action//modules/hub?ref=main"
  source                   = "../../../../monad-action/modules/integration"
  depends_on               = [aws_iam_openid_connect_provider.github]
  topology                 = module.topology
}

module "deployment" {
  # source = "github.com/bkeane/monad-action//modules/spoke?ref=main"
  source                   = "../../../../monad-action/modules/deployment"
  depends_on               = [aws_iam_openid_connect_provider.github]
  topology                 = module.topology
  api_gateway_ids          = toset([module.api_gateway.api_id])
  boundary_policy_document = module.boundary
  oidc_policy_document     = module.extended
}

resource "local_file" "integration_action" {
  content = module.topology.action.integration
  filename = "../../../.github/actions/integration/action.yaml"
}

resource "local_file" "deployment_action" {
  content = module.topology.action.deployment
  filename = "../../../.github/actions/deployment/action.yaml"
}

output "topology" {
  value = module.topology
}
