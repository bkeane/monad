FROM python:alpine
ARG SOURCE_DATE_EPOCH
COPY --from=public.ecr.aws/awsguru/aws-lambda-adapter:0.8.4 /lambda-adapter /opt/extensions/lambda-adapter
COPY requirements.txt requirements.txt
RUN pip install --no-cache-dir -r requirements.txt
COPY /src /src
ENV PORT=8090
ENV READINESS_CHECK_PATH=/health
ENV AWS_LWA_ERROR_STATUS_CODES=500-599
WORKDIR /src
ENTRYPOINT ["python", "main.py"]
