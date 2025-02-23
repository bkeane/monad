FROM python:alpine

COPY --from=public.ecr.aws/awsguru/aws-lambda-adapter:0.8.4 /lambda-adapter /opt/extensions/lambda-adapter

WORKDIR /src
COPY requirements.txt .
RUN pip install -r requirements.txt
COPY main.py .
COPY models.py .

ENV PORT=8090
ENV READINESS_CHECK_PATH=/health
ENTRYPOINT ["python", "main.py"]
