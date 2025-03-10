data "aws_iam_policy_document" "boundary" {
    statement {
        sid = "AllowIAMAccess"
        actions = [
            "iam:Get*",
            "iam:List*"
        ]
        resources = ["*"]
    }

    statement {
        sid = "AllowLambdaAccess"
        actions = [
            "lambda:List*",
            "lambda:Get*"
        ]
        resources = ["*"]
    }

    statement {
        sid = "AllowAPIGatewayAccess"
        actions = [
            "apigateway:GET"
        ]
        resources = ["*"]
    }

    statement {
        sid = "AllowCloudWatchAccess"
        actions = [
            "logs:Get*",
            "logs:List*"
        ]
        resources = ["*"]
    }

    statement {
        sid = "AllowEventBridgeAccess"
        actions = [
            "events:List*",
            "events:Describe*"
        ]
        resources = ["*"]
    }

    statement {
        sid = "AllowSSMParameterStoreAccess"
        actions = [
            "ssm:GetParameter",
            "ssm:GetParameterHistory",
            "ssm:GetParametersByPath",
            "ssm:GetParameters",
            "ssm:DescribeParameters"
        ]
        resources = ["*"]
    }
}

output "json" {
    value = data.aws_iam_policy_document.boundary.json
}

output "minified_json" {
    value = data.aws_iam_policy_document.boundary.json
}
