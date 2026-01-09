########################### Stage 0 ########################
ARG BUILD_IMAGE=golang:1.24.11
ARG BASE_IMAGE=ubuntu:24.04

FROM ${BUILD_IMAGE} AS builder

COPY ./server /go/src

ARG SERVER_VERSION=0.1.0
ARG GOPROXY_URL="http://goproxy.cn,direct"

ENV GOPROXY=${GOPROXY_URL}

RUN set -ex; \
    cd /go/src; \
    go mod download -x; \
    export flags="-X 'data-model/version.ServerVersion=${SERVER_VERSION}'"; \
    CGO_ENABLED=1 go build -ldflags "-s -w $flags" -o ./bin/data-model-server;

########################### Stage 1 ########################
# Copy working directory to the actual release docker images
FROM ${BASE_IMAGE} AS prod

ARG UID=1001
ARG GID=1001

RUN groupadd -r -g $GID adp \
    && useradd -r -u $UID -g $GID adp

COPY --from=builder --chown=$UID:$GID /go/src/bin /opt/data-model

ADD server/config/data-model-config.yaml /opt/data-model/config/
ADD server/locale /opt/data-model/locale

# 指定工作目录
WORKDIR /opt/data-model/

# Change user
USER $UID

# Expose port
EXPOSE 13020

# 指定启动命令
ENTRYPOINT [ "./data-model-server" ]
