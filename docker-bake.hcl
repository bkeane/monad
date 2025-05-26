group "default" {
  targets = ["build"]
}

variable "BRANCH" {
  description = "Branch name to use for caching"
}

variable "TAG" {
  description = "Image tag to use for output"
}

target "build" {
  context = "e2e/echo"
  platforms = ["linux/amd64", "linux/arm64"]

  output = [
    "type=image,name=${TAG},push=true"
  ]

  cache-to = [{
    type = "s3"
    region = "us-west-2"
    bucket = "kaixo-buildx-cache"
    prefix = "bkeane/monad/${BRANCH}"
    name = "echo"
    mode = "max"
  }]

  cache-from = [
    {
      type = "s3"
      region = "us-west-2"
      bucket = "kaixo-buildx-cache"
      prefix = "bkeane/monad/${BRANCH}"
      name = "echo"
    },
    {
      type = "s3"
      region = "us-west-2"
      bucket = "kaixo-buildx-cache"
      prefix = "bkeane/monad/main"
      name = "echo"
    }
  ]
}