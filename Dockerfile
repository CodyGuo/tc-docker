FROM golang:1.12-alpine as builder

ARG VERSION
COPY . /go/src/tc-docker
WORKDIR /go/src/tc-docker
RUN set -ex \
    && apk add --no-cache tzdata \
    && cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime \
    && export GO111MODULE="on" \
    && export GOPROXY=https://goproxy.io \
    && go install

FROM alpine
RUN set -ex \
    && apk add --no-cache tzdata iproute2 \
    && cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime \
    && echo "Asia/Shanghai" > /etc/timezone \
    && ln -sf /sbin/ip /usr/sbin/ip \
    && ln -sf /sbin/tc /usr/sbin/tc \
    && mkdir -p /var/run/netns
ENV DOCKER_HOST=unix:///var/run/docker.sock
ENV DOCKER_API_VERSION=1.40
WORKDIR /opt/app
COPY --from=builder /go/bin/tc-docker .
ENTRYPOINT ["/opt/app/tc-docker"]
