variable "TAG" {
  description = "Image tag to use for output"
}

target "build" {
  matrix = {
    arch = ["amd64", "arm64"]
  }

  name = "${arch}"
  context = "e2e/echo"
  platforms = ["linux/${arch}"]
  tag = ["${arch}"]
  load = true

  output = [
    "type=docker,name=echo-${arch}"
  ]

  cache-from = [{
    type = "s3"
    region = "us-west-2"
    bucket = "kaixo-buildx-cache"
    prefix = "bkeane/monad/echo/${arch}/"
    name = "echo"
  }]

  cache-to = [{
    type = "s3"
    region = "us-west-2"
    bucket = "kaixo-buildx-cache"
    prefix = "bkeane/monad/echo/${arch}/"
    name = "echo"
    mode = "max"
  }]
}

target "join" {
  context = "e2e/echo"
  platforms = ["linux/amd64", "linux/arm64"]
  tag = ["${TAG}"]
  load = true
  output = [
    "type=docker,name=echo"
  ]
}