FROM golang:1.25-alpine AS builder

WORKDIR /build

COPY go.work go.work
COPY pkg/ pkg/
COPY services/auth/ services/auth/
COPY services/finance/go.mod services/finance/go.mod
COPY services/finance/go.sum services/finance/go.sum
COPY services/gateway/go.mod services/gateway/go.mod
COPY services/gateway/go.sum services/gateway/go.sum

RUN cd services/auth && go build -o /auth-service .

FROM alpine:3.22

RUN apk add --no-cache ca-certificates
COPY --from=builder /auth-service /usr/local/bin/auth-service

EXPOSE 8081
ENTRYPOINT ["auth-service"]
