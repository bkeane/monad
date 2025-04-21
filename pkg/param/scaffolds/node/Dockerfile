FROM node:lts-alpine

COPY --from=public.ecr.aws/awsguru/aws-lambda-adapter:0.8.4 /lambda-adapter /opt/extensions/lambda-adapter

WORKDIR /var/task
COPY main.js .

ENV PORT=8090
ENV READINESS_CHECK_PATH=/health
ENTRYPOINT [ "node", "main.js" ]