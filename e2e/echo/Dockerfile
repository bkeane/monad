FROM python:alpine
COPY --from=public.ecr.aws/awsguru/aws-lambda-adapter:0.8.4 /lambda-adapter /opt/extensions/lambda-adapter
RUN mkdir -p /src
COPY requirements.txt /src/requirements.txt
RUN pip install --no-cache-dir -r /src/requirements.txt
COPY src/ src/
ENV PORT=8090
ENV READINESS_CHECK_PATH=/health
WORKDIR /src
ENTRYPOINT ["python", "/src/main.py"]
