FROM golang:1.23.3-alpine3.20 AS builder
WORKDIR /app
COPY /app .

RUN apk add --no-cache ca-certificates

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ms

FROM scratch
WORKDIR /app
COPY --from=builder /app/ms .
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
CMD [ "./ms" ]