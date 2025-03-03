output "workflow" {
    value = yamlencode({
        jobs = {
            publish = {
                runs-on = "ubuntu-latest"
                permissions = {
                    id-token = "write"
                    contents = "read"
                }
                matrix = {
                    chdir = []
                }
                steps = [
                    {
                        name = "Authenticate with AWS"
                        id = "assume-role"
                        uses = "aws-actions/configure-aws-credentials@v4"
                        with = {
                            role-to-assume = local.hub_account_role_arn
                            aws-region = data.aws_region.current.name
                        }
                    },
                    {
                        name = "Authenticate with ECR"
                        id = "docker-login"
                        uses = "aws-actions/amazon-ecr-login@v2"
                    },
                    {
                        name = "Checkout"
                        id = "checkout"
                        uses = "actions/checkout@v4"
                        with = {
                            fetch-depth = 0
                        }
                    },
                    {
                        name = "Install Monad"
                        id = "install-monad"
                        uses = "bkeane/monad-action@main"
                        with = {
                            version = "latest"
                        }
                    }
                ]
            }
        }
    })
}