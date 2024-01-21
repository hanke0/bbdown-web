# bbdown-web
A simple web application based on BBDown.

# How to install

Run manually:
```
AUTH_USER=set-your-username AUTH_PWD=set-your-password ./bbdown-web --addr '0.0.0.0:9280' --download ./
```

Run with docker:
```
docker run -v /downloads:/downloads -e AUTH_USER=set-your-username -e AUTH_PWD=set-your-password -e LISTEN_PORT=9280 googletranslate/bbdown-web:latest
```

Now you can open http://127.0.0.1:9080.
