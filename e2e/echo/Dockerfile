FROM python:alpine

COPY --from=public.ecr.aws/awsguru/aws-lambda-adapter:0.8.4 /lambda-adapter /opt/extensions/lambda-adapter

ADD requirements.txt requirements.txt
RUN pip install --no-cache-dir -r requirements.txt

COPY src/ src/
ENV PORT=8090
ENV READINESS_CHECK_PATH=/health
WORKDIR /src
ENTRYPOINT ["python", "/src/main.py"]
