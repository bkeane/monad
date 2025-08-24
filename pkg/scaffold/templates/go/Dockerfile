FROM golang:alpine AS build
WORKDIR /src
COPY main.go .
RUN go mod init headers
RUN go mod tidy
RUN GOOS=linux CGO_ENABLED=0 go build main.go

FROM scratch
COPY --from=public.ecr.aws/awsguru/aws-lambda-adapter:0.8.4 /lambda-adapter /opt/extensions/lambda-adapter
COPY --chmod=755 --from=build /src/main /var/task/main

ENV PORT=8090
ENV READINESS_CHECK_PATH=/health
ENTRYPOINT [ "/var/task/main" ]