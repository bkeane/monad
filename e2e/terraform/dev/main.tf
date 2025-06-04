locals {
    topology = data.terraform_remote_state.prod.outputs.topology
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

module "e2e_policy" {
  source = "../modules/e2e"
  depends_on = [aws_iam_openid_connect_provider.github]
  api_gateway_ids = [module.api_gateway.api_id]
}

module "monad_policy" {
  source = "../modules/monad"
  depends_on = [aws_iam_openid_connect_provider.github]
  git_repo_name = local.topology.git.repo
  ecr_repositories = local.topology.ecr_repositories
  api_gateway_ids = toset([module.api_gateway.api_id])
}

module "deploy" {
  source = "github.com/bkeane/stage/stage?ref=v0.1.0"
  depends_on = [aws_iam_openid_connect_provider.github]
  stage                    = "deploy"
  topology                 = local.topology
  policy_document          = module.monad_policy
}

module "e2e" {
  source = "github.com/bkeane/stage/stage?ref=v0.1.0"
  depends_on = [aws_iam_openid_connect_provider.github]
  stage                    = "e2e"
  topology                 = local.topology
  policy_document          = module.e2e_policy
}

output "topology" {
    value = data.terraform_remote_state.prod.outputs.topology
}