group "default" {
  targets = ["monad", "echo"]
}

variable "GIT_SHA" {
  description = "Git Sha"
  validation {
    condition     = GIT_SHA != null && GIT_SHA != ""
    error_message = "GIT_SHA must be provided and cannot be empty."
  }
}

variable "GIT_BRANCH" {
  description = "Git Branch"
  validation {
    condition     = GIT_BRANCH != null && GIT_BRANCH != ""
    error_message = "GIT_BRANCH must be provided and cannot be empty."
  }
}

target "monad" {
  context = "."
  description = "scratch image containing monad binary"
  platforms = ["linux/amd64", "linux/arm64"]
  load = true

  output = [
    "type=image,name=677771948337.dkr.ecr.us-west-2.amazonaws.com/bkeane/monad/cmd:${GIT_BRANCH}",
  ]
}

target "echo" {
  context = "e2e/echo"
  description = "scratch image containing e2e echo service"
  platforms = ["linux/amd64", "linux/arm64"]
  load = true

  output = [
    "type=image,name=677771948337.dkr.ecr.us-west-2.amazonaws.com/bkeane/monad/echo:${GIT_BRANCH}"
  ]
}