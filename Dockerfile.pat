FROM alpine:3.22

RUN apk update && \
  apk add --no-cache \
    bash \
    curl \
    jq \
    openssl

WORKDIR /app
COPY scripts/token.sh .

ENTRYPOINT ["/app/token.sh"]
