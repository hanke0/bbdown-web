FROM golang:1.19 as build

RUN apt-get update && apt-get install -y curl unzip

RUN curl -o /tmp/bbdown.zip -sSL https://github.com/nilaoda/BBDown/releases/download/1.5.4/BBDown_1.5.4_20221019_linux-x64.zip \
    && cd /tmp/ && unzip bbdown.zip

COPY . /app

RUN cd /app && go build -o /tmp/bbdown-web ./cmd/main.go

FROM linuxserver/ffmpeg

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
    LISTEN_ADDR=":9280" \
    DOWNLOAD="/downloads"

EXPOSE 9280
WORKDIR /app
HEALTHCHECK --interval=30s --timeout=30s --start-period=5s --retries=3 CMD [ "/app/healthy.sh" ]
ENTRYPOINT ["/app/start.sh"]
