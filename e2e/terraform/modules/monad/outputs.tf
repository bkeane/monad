output "json" {
    value = data.aws_iam_policy_document.deployment.json
}

output "minified_json" {
    value = data.aws_iam_policy_document.deployment.minified_json
}