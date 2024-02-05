# bbdown-web
A simple web application based on BBDown.

# How to install

Run manually:
```
AUTH_USER=set-your-username AUTH_PWD=set-your-password ./bbdown-web --addr '0.0.0.0:9280' --download ./ --bbdown-config ./bbdown.config
```

Run with docker:
```
docker run -v /downloads:/downloads -v /etc/bbdown.config:/app/bbdown.config -e AUTH_USER=set-your-username -e AUTH_PWD=set-your-password -e LISTEN_PORT=9280 -e PUID=1000 -e PGID=1000 googletranslate/bbdown-web:latest
```

Now you can open http://127.0.0.1:9080.
