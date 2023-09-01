FROM alpine

# proper reaping / signal forwarding init
RUN apk add --no-cache tini ca-certificates
ENTRYPOINT ["/sbin/tini", "--"]

ADD ./dist/linux/amd64/* /
