variable "api_gateway_ids" {
    type = list(string)
}

data "aws_caller_identity" "current" {}
data "aws_region" "current" {}

data "aws_iam_policy_document" "extended" {
    statement {
        sid = "AllowAPIGatewayInvoke"
        actions = [
            "execute-api:Invoke"
        ]
        resources = [
            for api_gateway_id in var.api_gateway_ids : "arn:aws:execute-api:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:${api_gateway_id}/*"
        ]
    }

    statement {
        sid = "AllowAPIGatewayDescribe" 
        actions = [
            "apigateway:GET"
        ]
        resources = ["*"]
    }

    statement {
        sid = "AllowEventBridgeAccess"
        actions = [
            "events:PutEvents"
        ]
        resources = [
            "arn:aws:events:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:event-bus/default",
        ]
    }

    statement {
        sid = "AllowSSMParameterStoreAccess"
        actions = [
            "ssm:GetParameter",
            "ssm:GetParameterHistory",
            "ssm:GetParametersByPath",
            "ssm:GetParameters",
            "ssm:DescribeParameters",
            "kms:Decrypt"
        ]
        resources = [
            "arn:aws:ssm:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:parameter/monad/*",
        ]
    }
}

output "json" {
    value = data.aws_iam_policy_document.extended.json
}

output "minified_json" {
    value = data.aws_iam_policy_document.extended.json
}
