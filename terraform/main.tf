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
}

data "aws_caller_identity" "current" {}
data "aws_region" "current" {}
data "corefunc_url_parse" "origin" {
  url = var.origin
}