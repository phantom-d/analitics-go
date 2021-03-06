FROM golang:1.17.6-alpine3.15 as builder

ARG TIME_ZONE="Europe/Moscow"

ENV TIME_ZONE=${TIME_ZONE} \
	TZ=${TIME_ZONE}

RUN apk update; \
    apk add --no-cache git gcc upx ca-certificates tzdata; \
    update-ca-certificates; \
    adduser -D -g '' appuser; \
    echo ${TIME_ZONE} > /etc/timezone

WORKDIR /app

ENV GO111MODULE=on

COPY . /app

RUN time go get -v -t ./...
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 time go build -v -a -ldflags="-w -s" -o programm ./cmd
RUN time upx --ultra-brute programm

FROM alpine:3.15.0

ARG TIME_ZONE="Europe/Moscow"

ENV TIME_ZONE=${TIME_ZONE} \
	TZ=${TIME_ZONE} \
    DB_TYPE=clickhouse \
	DB_HOST=127.0.0.1 \
    DB_PORT=9000 \
    DB_NAME=default \
    DB_USER="" \
    DB_PASS="" \
    DB_CERT="" \
    EXCHANGE_HOST=http://exchange.microk8s.fs.local \
    EXCHANGE_USER=analitics_test \
    EXCHANGE_PASSWORD=analitics_test

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /etc/timezone /etc/timezone
COPY --from=builder /usr/share/zoneinfo/${TIME_ZONE} /etc/localtime
COPY --from=builder /app/programm /app/programm
COPY --from=builder /app/migrations /app/migrations
COPY --from=builder /app/config.yaml /app/config.yaml

WORKDIR /app

USER appuser

EXPOSE ${HTTP_BIND}

ENTRYPOINT [ "/app/programm" ]

CMD [ "--migrate", "--daemon=watcher", "--config=/app/config.yaml", "--pid-dir=pids" ]
