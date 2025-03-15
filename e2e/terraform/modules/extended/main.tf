variable "account_id" {
    type = string
}

variable "region" {
    type = string
}

variable "api_gateway_ids" {
    type = list(string)
}

data "aws_iam_policy_document" "extended" {
    statement {
        sid = "AllowAPIGatewayInvoke"
        actions = [
            "execute-api:Invoke"
        ]
        resources = [
            for api_gateway_id in var.api_gateway_ids : "arn:aws:execute-api:${var.region}:${var.account_id}:${api_gateway_id}/*"
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
            "arn:aws:events:${var.region}:${var.account_id}:rule/monad-*",
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
            "arn:aws:ssm:${var.region}:${var.account_id}:parameter/monad/*",
        ]
    }
}

output "json" {
    value = data.aws_iam_policy_document.extended.json
}

output "minified_json" {
    value = data.aws_iam_policy_document.extended.json
}
