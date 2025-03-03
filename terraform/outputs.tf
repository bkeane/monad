output "role_arn" {
    value = aws_iam_role.github.arn
}

output "github_workflow_yaml" {
    value = yamlencode({
        jobs = {
            example = {
                runs-on = "ubuntu"
                permissions = {
                    id-token = "write"
                    contents = "read"
                }
                steps = [
                    {
                        name = "Retrieve AWS credentials"
                        uses = "aws-actions/configure-aws-credentials@v4"
                        with = {
                            role-to-assume = aws_iam_role.github.arn
                            aws-region = data.aws_region.current.name
                        }
                    },
                    {
                        name = "Checkout"
                        uses = "actions/checkout@v4"
                    }
                ]
            }
        }
    })
}