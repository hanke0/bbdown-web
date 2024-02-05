FROM golang:1.21 as build

RUN apt-get update && apt-get install -y curl unzip

RUN curl -o /tmp/bbdown.zip -sSL https://github.com/nilaoda/BBDown/releases/download/1.6.1/BBDown_1.6.1_20230818_linux-x64.zip \
    && cd /tmp/ && unzip bbdown.zip

COPY go.mod /app/go.mod
COPY cmd/* /app/cmd/
RUN cd /app && go build -o /tmp/bbdown-web ./cmd/main.go

FROM linuxserver/ffmpeg:latest

RUN mkdir -p /downloads /config \
    && apt-get update \
    && apt-get install -y curl gosu \
    && rm -rf /tmp/*

COPY --from=build /tmp/bbdown-web /tmp/BBDown /app/

COPY start.sh healthy.sh bbdown.config /app/

RUN chmod +x /app/BBDown \
    && chmod +x /app/bbdown-web \
    && chmod +x /app/start.sh \
    && chmod +x /app/healthy.sh \
    && useradd -d /app -u 1000 -M -U user

ENV BBDOWN=/app/BBDown \
    BBDOW_CONFIG=/app/bbdown.config \
    LISTEN_HOST="0.0.0.0" \
    LISTEN_PORT="9280" \
    DOWNLOAD="/downloads" \
    AUTH_USER="" \
    AUTH_PWD="" \
    PUID=1000 \
    PGID=1000

EXPOSE 9280
WORKDIR /app
HEALTHCHECK --interval=30s --timeout=30s --start-period=5s --retries=3 CMD [ "/app/healthy.sh" ]
LABEL version="1.6.1.2"
ENTRYPOINT ["/app/start.sh"]

LABEL bbdown-version="1.6.1"
LABEL author="hanke"
