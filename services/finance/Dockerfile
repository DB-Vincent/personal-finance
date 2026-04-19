FROM golang:1.25-alpine AS builder

WORKDIR /build

COPY go.work go.work
COPY pkg/ pkg/
COPY services/finance/ services/finance/
COPY services/auth/go.mod services/auth/go.mod
COPY services/auth/go.sum services/auth/go.sum
COPY services/gateway/go.mod services/gateway/go.mod
COPY services/gateway/go.sum services/gateway/go.sum

RUN cd services/finance && go build -o /finance-service .

FROM alpine:3.22

RUN apk add --no-cache ca-certificates
COPY --from=builder /finance-service /usr/local/bin/finance-service

EXPOSE 8082
ENTRYPOINT ["finance-service"]
