FROM alpine
RUN apk add --no-cache jq curl
COPY --chmod=+x main.sh /var/task/main.sh
ENTRYPOINT ["/var/task/main.sh"]