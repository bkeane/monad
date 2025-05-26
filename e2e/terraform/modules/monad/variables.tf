variable "name" {
    description = "The name of the policy and role"
    type = string
    default = "monad-action"
}

variable "git_repo_name" {
    description = "The name of the git repository"
    type = string
}

variable "repositories" {
    description = "The ecr repositories under management"
    type = list(string)
}

variable "api_gateway_ids" {
    description = "The api gateway ids under management"
    type = list(string)
}

variable "boundary_policy_arn" {
    description = "The boundary policy arn for roles under management"
    type = string
    default = null
}