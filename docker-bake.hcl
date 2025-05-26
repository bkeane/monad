group "default" {
  targets = ["build"]
}

variable "TAG" {
  description = "Image tag to use for output"
}

target "build_arch" {
  matrix = {
    arch = ["amd64", "arm64"]
  }

  name = "${arch}"
  context = "e2e/echo"
  platforms = ["linux/${arch}"]
  tag = ["${arch}"]

  output = [
    "type=image,name=${TAG}",
  ]

  cache-from = [{
    type = "s3"
    region = "us-west-2"
    bucket = "kaixo-buildx-cache"
    prefix = "${arch}/"
    name = "echo"
  }]

  cache-to = [{
    type = "s3"
    region = "us-west-2"
    bucket = "kaixo-buildx-cache"
    prefix = "${arch}/"
    name = "echo"
    mode = "max"
  }]
}

target "build" {
  context = "e2e/echo"
  platforms = ["linux/amd64", "linux/arm64"]
  tag = [TAG]
  output = [
    "type=image,name=${TAG},push=true",
  ]
}