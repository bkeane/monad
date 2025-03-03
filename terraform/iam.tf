
#
# GitHub OIDC Provider
#

resource "aws_iam_openid_connect_provider" "github" {
  count          = var.create_oidc_provider ? 1 : 0
  url            = "https://token.actions.githubusercontent.com"
  client_id_list = ["sts.amazonaws.com"]
  # Hex-encoded SHA-1 hash of the X.509 domain certificate
  # https://github.blog/changelog/2023-06-27-github-actions-update-on-oidc-integration-with-aws/
  thumbprint_list = [
    "6938fd4d98bab03faadb97b34396831e3780aea1",
    "1c58a3a8518e8759bf075b76b750d4f2df264fcd"
  ]
}

data "aws_iam_openid_connect_provider" "github" {
  count = var.create_oidc_provider ? 0 : 1
  arn   = "arn:aws:iam::${var.ecr_hub_account_id}:oidc-provider/token.actions.githubusercontent.com"
}

#
# GitHub Actions Role
#

resource "aws_iam_role" "github" {
  name                  = "${local.prefix}-oidc-role"
  description           = "used by ${var.origin} github actions"
  assume_role_policy    = data.aws_iam_policy_document.trust.json
  force_detach_policies = true
}

data "aws_iam_policy_document" "trust" {
  statement {
    principals {
      type        = "Federated"
      identifiers = var.create_oidc_provider ? [aws_iam_openid_connect_provider.github[0].arn] : [data.aws_iam_openid_connect_provider.github[0].arn]
    }

    actions = ["sts:AssumeRoleWithWebIdentity"]

    condition {
      test     = "StringEquals"
      variable = "token.actions.githubusercontent.com:aud"
      values   = ["sts.amazonaws.com"]
    }

    condition {
      test     = "StringLike"
      variable = "token.actions.githubusercontent.com:sub"
      values = [
          "repo:${local.repo_path}"
      ]
    }
  }
}

resource "aws_iam_role_policy_attachment" "github" {
  role       = aws_iam_role.github.name
  policy_arn = aws_iam_policy.github.arn
}

#
# GitHub Actions Policy
#

resource "aws_iam_policy" "github" {
  name        = "${local.prefix}-oidc-policy"
  description = "used by ${var.origin} github actions"
  policy      = data.aws_iam_policy_document.monad.json
}


data "aws_iam_policy_document" "monad" {
  statement {
    sid       = "AllowEcrLogin"
    effect    = "Allow"
    resources = ["*"]
    actions   = ["ecr:GetAuthorizationToken"]
  }

  statement {
    sid    = "AllowEcrRepositoryAccess"
    effect = "Allow"
    actions = [
      "ecr:*",
    ]
    resources = [
      for image_path in local.images :
        "arn:aws:ecr:${data.aws_region.current.name}:${var.ecr_hub_account_id}:repository/${image_path}"
    ]
  }

  statement {
    sid    = "AllowIAMAccess"
    effect = "Allow"
    actions = [
      "iam:*"
    ]
    resources = ["*"]
    # condition {
    #   test     = "StringEquals"
    #   variable = "iam:PermissionsBoundary"
    #   values   = [aws_iam_policy.github_boundary.arn]
    # }
  }

  statement {
    sid    = "AllowLambdaAccess"
    effect = "Allow"
    actions = [
      "lambda:*"
    ]
    resources = ["*"]
  }

  statement {
    sid    = "AllowApiGatewayV2Access"
    effect = "Allow"
    actions = [
      "apigateway:*"
    ]
    resources = [
      "arn:aws:apigateway:${data.aws_region.current.name}::/apis",
      "arn:aws:apigateway:${data.aws_region.current.name}::/apis/*",
      "arn:aws:apigateway:${data.aws_region.current.name}::/tags/*"
    ]
  }

  statement {
    sid    = "AllowCloudWatchAccess"
    effect = "Allow"
    actions = [
      "cloudwatch:*"
    ]
    resources = ["*"]
  }

  statement {
    sid = "AllowEventBridgeAccess"
    effect = "Allow"
    actions = [
      "events:*"
    ]
    resources = ["*"]
  }
}

#
# GitHub Actions Boundary Policy
#

# resource "aws_iam_policy" "github_boundary" {
#   name        = "${var.prefix}-oidc-boundary-policy"
#   description = "used by monad in github actions"
#   policy      = data.aws_iam_policy_document.github_boundary.json
# }

