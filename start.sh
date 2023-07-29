#!/bin/bash

extra=()
if [ -n "$BBDOWNOPTION" ]; then
    extra=(--bbdown-option "$BBDOWNOPTION")
fi

exec /app/bbdown-web "${extra[@]}" -bbdown "$BBDOWN" -addr "$LISTEN_HOST:$LISTEN_PORT" -download "$DOWNLOAD"
