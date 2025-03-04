terraform {
  required_providers {
    corefunc = {
      source  = "northwood-labs/corefunc"
      version = "~> 1.0"
    }
  }
}

locals {
  is_hub = var.ecr_hub_account_id == data.aws_caller_identity.current.account_id

  ecr_hub_account_id = var.ecr_hub_account_id != null ? var.ecr_hub_account_id : data.aws_caller_identity.current.account_id
  ecr_hub_account_region = var.ecr_hub_account_region != null ? var.ecr_hub_account_region : data.aws_region.current.name
  ecr_hub_account_role_arn = "arn:aws:iam::${local.ecr_hub_account_id}:role/${local.prefix}-oidc-role"
  
  ecr_spoke_account_role_arns = [
    for account_id in var.ecr_spoke_account_ids :
    "arn:aws:iam::${account_id}:role/${local.prefix}-oidc-role"
  ]

  repo_path  = replace(data.corefunc_url_parse.origin.path, ".git", "")
  repo_parts = compact(split("/", local.repo_path))
  repo_owner = local.repo_parts[0]
  repo_name  = local.repo_parts[1]

  prefix           = "${local.repo_owner}-${local.repo_name}"
  allowed_branches = "*"

  images = toset([
    for path, service in var.services :
      "${local.repo_owner}/${local.repo_name}/${basename(path)}"
  ])


}

data "aws_caller_identity" "current" {}
data "aws_region" "current" {}
data "corefunc_url_parse" "origin" {
  url = var.origin
}
