group "default" {
  targets = ["echo"]
}

variable "CACHE_PREFIX" {
  default = "local"
}

variable "ECR_TAG" {
  description = "Image tag to use for output"
}

variable "SOURCE_DATE_EPOCH" {
  default = "0"
}

target "echo" {
  context = "e2e/echo"
  platforms = ["linux/amd64", "linux/arm64"]
  output = [
    "type=registry,name=${ECR_TAG},rewrite-timestamp=true",
  ]
  cache-from = ["type=s3,region=us-west-2,bucket=kaixo-buildx-cache,prefix=${CACHE_PREFIX}/,name=echo"]
  cache-to = ["type=s3,region=us-west-2,bucket=kaixo-buildx-cache,prefix=${CACHE_PREFIX}/,name=echo,mode=max"]
  args = {
    SOURCE_DATE_EPOCH = "${SOURCE_DATE_EPOCH}"
  }
}