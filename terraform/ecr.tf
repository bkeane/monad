resource "aws_ecr_repository" "services" {
    for_each = local.is_hub ? toset(local.images) : toset([])
    name = each.value
    image_tag_mutability = var.mutable ? "MUTABLE" : "IMMUTABLE"
}

resource "aws_ecr_repository_policy" "cross_account_access" {
    for_each = aws_ecr_repository.services
    repository = each.value.name
    policy = jsonencode({
        Version = "2012-10-17"
        Statement = [
            {
                Sid = "CrossAccountPermission"
                Effect = "Allow"
                Principal = {
                    AWS = concat([
                        for account_id in var.ecr_spoke_account_ids:
                            "arn:aws:iam::${account_id}:root"
                    ])
                }
                Action = [
                    "ecr:BatchGetImage",
                    "ecr:GetDownloadUrlForLayer"
                ]
            },
            {
                Sid = "LambdaECRImageRetrievalPolicy"
                Effect = "Allow"
                Action = [
                    "ecr:BatchGetImage",
                    "ecr:GetDownloadUrlForLayer"
                ]
                Principal = {
                    Service = "lambda.amazonaws.com"
                }
                Condition = {
                    StringLike = {
                        "aws:sourceARN": concat([
                            for account_id in var.ecr_spoke_account_ids: 
                                "arn:aws:lambda:${data.aws_region.current.name}:${account_id}:function:*"
                        ])
                    }
                }
            }
        ]
    })
}