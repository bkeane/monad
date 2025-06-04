group "default" {
  targets = ["build"]
}

variable "BRANCH" {
  description = "Branch name"
  required = true
}

target "build" {
  context = "e2e/echo"
  platforms = ["linux/amd64", "linux/arm64"]

  output = [
    "type=image,name=677771948337.dkr.ecr.us-west-2.amazonaws.com/bkeane/monad/echo:${BRANCH},push=true"
  ]

  cache-to = [{
    type = "s3"
    region = "us-west-2"
    bucket = "kaixo-buildx-cache"
    prefix = "bkeane/monad/echo/${BRANCH}/"
    name = "echo"
    mode = "max"
  }]

  cache-from = [
    {
      type = "s3"
      region = "us-west-2"
      bucket = "kaixo-buildx-cache"
      prefix = "bkeane/monad/echo/${BRANCH}/"
      name = "echo"
    },
    {
      type = "s3"
      region = "us-west-2"
      bucket = "kaixo-buildx-cache"
      prefix = "bkeane/monad/echo/main/"
      name = "echo"
    }
  ]
}