FROM golang:1.19 as build

RUN apt-get update && apt-get install -y curl unzip

RUN curl -o /tmp/bbdown.zip -sSL https://github.com/nilaoda/BBDown/releases/download/1.6.0/BBDown_1.6.0_20230715_linux-x64.zip \
    && cd /tmp/ && unzip bbdown.zip

COPY go.mod /app/go.mod
COPY cmd/* /app/cmd/
RUN cd /app && go build -o /tmp/bbdown-web ./cmd/main.go

FROM linuxserver/ffmpeg:latest

RUN mkdir -p /downloads /config \
    && apt-get update \
    && apt-get install -y curl \
    && rm -rf /tmp/*

COPY --from=build /tmp/bbdown-web /tmp/BBDown /app/

COPY start.sh healthy.sh /app/

RUN chmod +x /app/BBDown \
    && chmod +x /app/bbdown-web \
    && chmod +x /app/start.sh \
    && chmod +x /app/healthy.sh

ENV BBDOWN=./BBDown \
    BBDOWNOPTION= \
    LISTEN_HOST="0.0.0.0" \
    LISTEN_PORT="9280" \
    DOWNLOAD="/downloads" \
    AUTH_USER="admin" \
    AUTH_PWD="admin"

EXPOSE 9280
WORKDIR /app
HEALTHCHECK --interval=30s --timeout=30s --start-period=5s --retries=3 CMD [ "/app/healthy.sh" ]
ENTRYPOINT ["/app/start.sh"]
