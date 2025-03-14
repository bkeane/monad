# syntax=docker/dockerfile:1
FROM --platform=$BUILDPLATFORM golang:alpine AS build
ARG TARGETOS
ARG TARGETARCH
WORKDIR /src
COPY . .
RUN GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -o monad ./cmd/monad

FROM alpine AS shellspec
RUN apk add --no-cache shellspec aws-cli git curl jq bash uuidgen
ENTRYPOINT ["/bin/bash", "-c", "shellspec --chdir /src/e2e"]

FROM scratch
COPY --from=build --chmod=755 /src/monad /monad
ENTRYPOINT ["/monad"]