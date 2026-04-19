FROM golang:1.25-alpine AS builder

WORKDIR /build

COPY go.work go.work
COPY pkg/ pkg/
COPY services/gateway/ services/gateway/
COPY services/auth/go.mod services/auth/go.mod
COPY services/auth/go.sum services/auth/go.sum

RUN cd services/gateway && go build -o /gateway .

FROM alpine:3.22

RUN apk add --no-cache ca-certificates
COPY --from=builder /gateway /usr/local/bin/gateway

EXPOSE 8080
ENTRYPOINT ["gateway"]
