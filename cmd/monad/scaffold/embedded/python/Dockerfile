FROM python:alpine

COPY --from=public.ecr.aws/awsguru/aws-lambda-adapter:0.8.4 /lambda-adapter /opt/extensions/lambda-adapter

WORKDIR /src
COPY main.py .

ENV PORT=8090
ENV READINESS_CHECK_PATH=/health
ENTRYPOINT ["python", "main.py"]
