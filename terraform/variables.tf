variable "origin" {
  type = string
  description = "https origin of the github repo"
}

variable "services" {
  type = map(object({
    global_args = optional(string, "")
    encode_args = optional(string, "")
    deploy_args = optional(string, "")
    destroy_args = optional(string, "")
  }))
  description = "service definitions"
  default = {}
}

variable "mutable" {
  type = bool
  description = "whether ECR repository image tags are mutable"
  default = true
}

variable "ecr_hub_account_id" {
  type = string
  description = "The ECR hub account ID"
}

variable "ecr_hub_account_region" {
  type = string
  description = "The ECR hub account region"
  default = null
}

variable "ecr_spoke_account_ids" {
  type = set(string)
  description = "The ECR spoke account IDs"
}

variable "create_oidc_provider" {
  type = bool
  description = "Whether to create the OIDC provider or lookup existing"
  default = false
}

# variable "boundary_policy" {
#   type = string
#   description = "The boundary policy ARN"
# }