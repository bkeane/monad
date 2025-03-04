provider "aws" {
  region = "us-west-2"
  profile = "dev.kaixo.io"
}

terraform {
  backend "s3" {
    bucket  = "kaixo-dev-tofu"
    key     = "testing/terraform.tfstate"
    region  = "us-west-2"
    encrypt = true
  }
}