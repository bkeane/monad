# group "default" {
#   targets = ["echo"]
# }

variable "BRANCH" {
  description = "Branch name"
  required = true
}

target "monad" {
  context = "."
  description = "scratch image containing monad binary"
  platforms = ["linux/amd64", "linux/arm64"]
  load = true

  output = [
    "type=image,name=677771948337.dkr.ecr.us-west-2.amazonaws.com/bkeane/monad/cmd:${BRANCH}",
  ]
}

target "echo" {
  context = "e2e/echo"
  description = "scratch image containing e2e echo service"
  platforms = ["linux/amd64", "linux/arm64"]
  load = true

  output = [
    "type=image,name=677771948337.dkr.ecr.us-west-2.amazonaws.com/bkeane/monad/echo:${BRANCH}"
  ]
}