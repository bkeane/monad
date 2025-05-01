resource "aws_s3_bucket" "cache" {
  bucket = "kaixo-buildx-cache"
}

resource "aws_s3_bucket_policy" "cache" {
  bucket = aws_s3_bucket.cache.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid       = "AllowFullAccess"
        Effect    = "Allow"
        Principal = {
          AWS = module.topology.oidc.integration_role_arn
        }
        Action = [
          "s3:*"
        ]
        Resource = [
          aws_s3_bucket.cache.arn,
          "${aws_s3_bucket.cache.arn}/*"
        ]
      }
    ]
  })
}
