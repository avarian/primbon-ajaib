FROM golang:1.18-alpine AS builder

RUN apk --update --no-cache add \
    build-base gcc linux-headers git && \
    mkdir /build

COPY . /build
WORKDIR /build

RUN make linux

FROM alpine:latest

ARG TIMEZONE=Asia/Jakarta

RUN apk --update --no-cache add \
    bash \
    curl \
    tzdata \
    ca-certificates && \
    ln -sf "/usr/share/zoneinfo/${TIMEZONE}" /etc/localtime && \
    echo ${TIMEZONE} > /etc/timezone && \
    curl -sSo /usr/local/bin/wait-for-it.sh "https://raw.githubusercontent.com/vishnubob/wait-for-it/master/wait-for-it.sh" && \
    chmod +x /usr/local/bin/wait-for-it.sh && \
    apk del curl && \
    rm -rf /var/cache/apk/*

ENV DEFAULT_TZ ${TIMEZONE}
ENV LC_ALL en_US.UTF-8
ENV LANG en_US.UTF-8

COPY --from=builder /build/primbon-ajaib-backend_linux_amd64 /app/primbon-ajaib-backend
COPY primbon-ajaib-backend.yml /app/

WORKDIR /app
EXPOSE 8080

ENTRYPOINT [ "/app/primbon-ajaib-backend" ]
CMD [ "serve" ]
