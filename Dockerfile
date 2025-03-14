# syntax=docker/dockerfile:1
FROM --platform=$BUILDPLATFORM golang:alpine AS build
ARG TARGETOS
ARG TARGETARCH
RUN apk add --no-cache shellspec just
WORKDIR /src
COPY . .
RUN GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -o monad ./cmd/monad

FROM alpine AS extra
RUN apk add --no-cache shellspec just

FROM scratch
COPY --from=extra /usr/bin/just /just
COPY --from=extra /usr/bin/shellspec /shellspec
COPY --from=build --chmod=755 /src/monad /monad
ENTRYPOINT ["/monad"]