FROM python:alpine

# COPY --from=public.ecr.aws/awsguru/aws-lambda-adapter:0.8.4 /lambda-adapter /opt/extensions/lambda-adapter

# RUN mkdir -p /src
WORKDIR /src
COPY requirements.txt requirements.txt

# RUN --mount=type=cache,target=/root/.cache/pip,id=pip-cache \
#     pip install --cache-dir /root/.cache/pip -r requirements.txt
# COPY src/ .

# ENV PORT=8090
# ENV READINESS_CHECK_PATH=/health

# ENTRYPOINT ["python", "main.py"]
