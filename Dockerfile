FROM alpine:3.4

# Hack alpine into thinking that muslc lib is Glibc lib.
# Since they are compatible, the program will not notice any difference.
# Source: http://stackoverflow.com/a/35613430/3461549
RUN mkdir /lib64 && ln -s /lib/libc.musl-x86_64.so.1 /lib64/ld-linux-x86-64.so.2

COPY synpost_stats /usr/local/bin/

CMD ["synpost_stats"]
