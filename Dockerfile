FROM golang:1.24-alpine3.22

WORKDIR /app

RUN apk add --no-cache openssl bash docker-cli
RUN adduser -S mountain-hawk && \
    addgroup docker && \
    adduser mountain-hawk docker

RUN go install github.com/github/github-mcp-server/cmd/mcpcurl@latest && \
  which mcpcurl

COPY . ./

RUN go mod download && \
  go build -o mountain-hawk ./cmd/cli && \
  go clean -cache -modcache

ENTRYPOINT ["/app/mountain-hawk"]
