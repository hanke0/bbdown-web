#!/bin/bash

exec /app/bbdown-web -bbdown "$BBDOWN" -addr "$LISTEN_ADDR" -download "$DOWNLOAD"
