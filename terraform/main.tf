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

    repo_path = replace(data.corefunc_url_parse.origin.path, ".git", "")
    repo_parts = compact(split("/", local.repo_path))
    repo_owner = local.repo_parts[0]
    repo_name = local.repo_parts[1]
    prefix = "${local.repo_owner}-${local.repo_name}"

    images = [
        for service in var.services:
            "${local.repo_owner}/${local.repo_name}/${service}"
    ]

    hub_account_role_arn = "arn:aws:iam::${var.ecr_hub_account_id}:role/${local.prefix}-oidc-role"
    spoke_account_role_arns = [
        for account_id in var.ecr_spoke_account_ids:
            "arn:aws:iam::${account_id}:role/${local.prefix}-oidc-role"
    ]
}

data "aws_caller_identity" "current" {}
data "aws_region" "current" {}
data "corefunc_url_parse" "origin" {
  url = var.origin
}