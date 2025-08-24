locals {
    account_id = data.aws_caller_identity.current.account_id
    resource_name_wildcard = "${var.git_repo_name}-*-*"
    resource_path_wildcard = "${var.git_repo_name}/*/*"
}

data "aws_caller_identity" "current" {}

data "aws_iam_policy_document" "deployment" {
  statement {
    sid    = "AllowEcrRegistryLogin"
    effect = "Allow"
    actions = [
      "ecr:DescribeRegistry",
      "ecr:GetAuthorizationToken",
      "ecr:GetDownloadUrlForLayer",
      "ecr:BatchGetImage",
      "ecr:BatchCheckLayerAvailability",
      "ecr:DescribeRepositories",
      "ecr:ListImages"
    ]
    resources = ["*"]
  }

  statement {
    sid    = "AllowEcrLambda"
    effect = "Allow"
    actions = [
      "ecr:SetRepositoryPolicy",
      "ecr:GetRepositoryPolicy",
    ]

    resources = [
        for repo in var.ecr_repositories : repo.arn
    ]
  }

  statement {
    // DENY the OIDC role the ability to assume the roles it creates.
    sid    = "DenyOIDCChaining"
    effect = "Deny"
    actions = [
      "sts:AssumeRole",
    ]

    resources = [
      "*"
    ]
  }

  dynamic "statement" {
    for_each = var.boundary_policy_arn != null ? [1] : []
    content {
      sid    = "DenyBoundaryPolicyDeletion"
      effect = "Deny"
      actions = [
        "iam:DeletePolicy",
        "iam:DeletePolicyVersion"
      ]
      resources = [
        var.boundary_policy_arn
      ]
    }
  }

  dynamic "statement" {
    for_each = var.boundary_policy_arn != null ? [1] : []
    content {
      sid       = "DenyRoleCreateWithoutBoundary"
      effect    = "Deny"
      actions   = ["iam:CreateRole"]
      resources = ["arn:aws:iam::${local.account_id}:role/${local.resource_name_wildcard}"]
      condition {
        test     = "StringNotEquals"
        variable = "iam:PermissionsBoundary"
        values   = [var.boundary_policy_arn]
      }
    }
  }

  statement {
    sid    = "AllowEniRoleWrite"
    effect = "Allow"
    actions = [
      "iam:*"
    ]
    resources = [
      "arn:aws:iam::${local.account_id}:role/AWSLambdaVPCAccessExecutionRole",
      "arn:aws:iam::aws:policy/service-role/AWSLambdaVPCAccessExecutionRole",
    ]
  }

  statement {
    sid    = "AllowIamWrite"
    effect = "Allow"
    actions = [
      "iam:*"
    ]
    resources = [
      "arn:aws:iam::${local.account_id}:policy/${local.resource_name_wildcard}",
      "arn:aws:iam::${local.account_id}:role/${local.resource_name_wildcard}",
    ]
  }

  statement {
    sid    = "AllowLambdaWrite"
    effect = "Allow"
    actions = [
      "lambda:*"
    ]
    resources = [
      "arn:aws:lambda:*:${local.account_id}:function:${local.resource_name_wildcard}",
    ]
  }

  statement {
    sid    = "AllowApiGatewayV2Read"
    effect = "Allow"
    actions = [
      "apigateway:GET"
    ]
    resources = ["*"]
  }

  statement {
    sid    = "AllowApiGatewayV2Write"
    effect = "Allow"
    actions = [
      "apigateway:*"
    ]
    resources = flatten([
      for id in var.api_gateway_ids : [
        "arn:aws:apigateway:*::/apis/${id}",
        "arn:aws:apigateway:*::/apis/${id}/*",
        "arn:aws:apigateway:*::/tags/${id}",
      ]
    ])
  }

  statement {
    sid    = "AllowCloudWatchWrite"
    effect = "Allow"
    actions = [
      "logs:*"
    ]
    resources = [
      "arn:aws:logs:*:${local.account_id}:log-group:/aws/lambda/${local.resource_name_wildcard}",
      "arn:aws:logs:*:${local.account_id}:log-group:/aws/lambda/${local.resource_name_wildcard}:*",
    ]
  }

  statement {
    sid    = "AllowEventBridgeRead"
    effect = "Allow"
    actions = [
      "events:List*",
      "events:Describe*",
    ]
    resources = ["*"]
  }

  statement {
    sid    = "AllowEventBridgeWrite"
    effect = "Allow"
    actions = [
      "events:*"
    ]
    resources = [
      "arn:aws:events:*:${local.account_id}:rule/${local.resource_name_wildcard}"
    ]
  }

  statement {
    sid = "AllowVPCRead"
    effect = "Allow"
    actions = [
      "ec2:DescribeSecurityGroups",
      "ec2:DescribeSubnets",
      "ec2:DescribeVpcs"
    ]
    resources = ["*"]
  }
}
