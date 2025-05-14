group "default" {
  targets = ["echo"]
}

variable "TAG" {
  description = "Image tag to use for output"
}

variable "EPOCH" {
  default = "0"
}

target "echo" {
  context = "e2e/echo"
  platforms = ["linux/amd64", "linux/arm64"]
  tag = [TAG]
  
  output = [
    "type=image,name=${TAG},rewrite-timestamp=true",
    "type=docker,name=${TAG}"
  ]

  cache-from = [{
    type = "s3"
    region = "us-west-2"
    bucket = "kaixo-buildx-cache"
    name = "echo"
  }]

  cache-to = [{
    type = "s3"
    region = "us-west-2"
    bucket = "kaixo-buildx-cache"
    name = "echo"
    mode = "max"
  }]

  args = {
    SOURCE_DATE_EPOCH = "${EPOCH}"
  }
}