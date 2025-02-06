FROM --platform=$BUILDPLATFORM golang:alpine AS build
ARG TARGETOS
ARG TARGETARCH
WORKDIR /build
COPY . .
RUN GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -o monad cmd/monad/main.go

FROM alpine
WORKDIR /var/task
COPY --from=public.ecr.aws/awsguru/aws-lambda-adapter:0.8.4 /lambda-adapter /opt/extensions/lambda-adapter
COPY --from=build --chmod=755 /build/monad /var/task/monad

ENV PORT=8090
ENV READINESS_CHECK_PATH=/health
ENTRYPOINT ["./monad"]
CMD ["listen"]
