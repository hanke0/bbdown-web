#!/bin/bash

exec /app/bbdown-web -bbdown "$BBDOWN" -addr "$LISTEN_HOST:$LISTEN_PORT" -download "$DOWNLOAD"
