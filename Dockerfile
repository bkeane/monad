# syntax=docker/dockerfile:1
FROM --platform=$BUILDPLATFORM golang:alpine AS build
ARG TARGETOS
ARG TARGETARCH
WORKDIR /src

COPY go.mod go.sum ./
RUN GOOS=${TARGETOS} GOARCH=${TARGETARCH} go mod download

COPY cmd/ cmd/
COPY pkg/ pkg/
COPY internal/ internal/
RUN GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -o monad ./cmd/monad

FROM scratch
COPY --from=build --chmod=755 /src/monad /monad
ENTRYPOINT ["/monad"]