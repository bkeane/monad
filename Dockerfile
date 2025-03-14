# syntax=docker/dockerfile:1
FROM --platform=$BUILDPLATFORM golang:alpine AS build
ARG TARGETOS
ARG TARGETARCH
RUN apk add --no-cache shellspec just
WORKDIR /src
COPY . .
RUN GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -o monad ./cmd/monad

FROM alpine AS shellspec
RUN apk add --no-cache shellspec aws-cli

FROM scratch
COPY --from=build --chmod=755 /src/monad /monad
ENTRYPOINT ["/monad"]