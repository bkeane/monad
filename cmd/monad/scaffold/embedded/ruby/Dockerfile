FROM ruby:alpine

COPY --from=public.ecr.aws/awsguru/aws-lambda-adapter:0.8.4 /lambda-adapter /opt/extensions/lambda-adapter

WORKDIR /var/task
RUN gem install webrick --no-document
COPY main.rb .

ENV PORT=8090
ENV READINESS_CHECK_PATH=/health
ENTRYPOINT [ "ruby", "main.rb" ]
