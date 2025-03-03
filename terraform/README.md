

## Providers

| Name | Version |
|------|---------|
| <a name="provider_aws"></a> [aws](#provider\_aws) | n/a |

## Modules

No modules.

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| <a name="input_create_oidc_provider"></a> [create\_oidc\_provider](#input\_create\_oidc\_provider) | Whether to create the OIDC provider or lookup existing | `bool` | `false` | no |
| <a name="input_ecr_hub_account_id"></a> [ecr\_hub\_account\_id](#input\_ecr\_hub\_account\_id) | The ECR hub account ID | `string` | n/a | yes |
| <a name="input_ecr_repository_paths"></a> [ecr\_repository\_paths](#input\_ecr\_repository\_paths) | The ECR repository paths | `set(string)` | n/a | yes |
| <a name="input_ecr_spoke_account_ids"></a> [ecr\_spoke\_account\_ids](#input\_ecr\_spoke\_account\_ids) | The ECR spoke account IDs | `set(string)` | n/a | yes |
| <a name="input_organization"></a> [organization](#input\_organization) | The github organization name | `string` | n/a | yes |
| <a name="input_prefix"></a> [prefix](#input\_prefix) | prefix for resource names | `string` | `"monad"` | no |

## Outputs

| Name | Description |
|------|-------------|
| <a name="output_github_workflow_yaml"></a> [github\_workflow\_yaml](#output\_github\_workflow\_yaml) | n/a |
| <a name="output_role_arn"></a> [role\_arn](#output\_role\_arn) | n/a |
