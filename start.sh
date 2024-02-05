#!/bin/sh

set -e

groupmod -g "$PGID" --non-unique user
usermod -u "$PUID" --non-unique user
chown -R "${PUID}:${PGID}" /app
exec gosu "${PUID}:${PGID}" /app/bbdown-web \
    -bbdown "$BBDOWN" \
    -addr "$LISTEN_HOST:$LISTEN_PORT" \
    --bbdown-config "$BBDOW_CONFIG" \
    -download "$DOWNLOAD"
