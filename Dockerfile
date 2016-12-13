FROM alpine:3.4

COPY synpost_stats /usr/local/bin/

CMD ["synpost_stats"]
