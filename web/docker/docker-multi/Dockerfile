ARG BASE_IMAGE=acr.aishu.cn/public/nginx:1.22.1-alpine
ARG VEGA_BUILD_IMAGE=acr.aishu.cn/public/node:22.18.0
ARG FLOW_BUILD_IMAGE=acr.aishu.cn/public/node:18-pnpm-7
# 默认使用 linux/amd64，可通过 --build-arg TARGETPLATFORM=linux/arm64 覆盖
ARG TARGETPLATFORM=linux/amd64

# ========== 第一阶段：构建 vega-web ==========
FROM --platform=${TARGETPLATFORM} ${VEGA_BUILD_IMAGE} AS vega-builder

COPY ./web /vega-web/
COPY ./web/.[^.]* /vega-web/

RUN set -ex; \
    cd /vega-web/; \
    npm install -g pnpm; \
    pnpm install; \
    pnpm run build;

# ========== 第二阶段：构建 flow-web ==========
FROM --platform=${TARGETPLATFORM} ${FLOW_BUILD_IMAGE} AS flow-builder

COPY ./dataflow/dia-flow-web /flow-web/

RUN set -ex; \
    cd /flow-web/; \
    echo "=== Installing dependencies ==="; \
    pnpm install --no-frozen-lockfile --registry https://registry.npmmirror.com/ --fetch-timeout 100000 --fetch-retries 5

RUN set -ex; \
    cd /flow-web/; \
    echo "=== Generating icons ==="; \
    pnpm --filter @applet/icons run g

RUN set -ex; \
    cd /flow-web/; \
    echo "=== Building dependencies ==="; \
    pnpm --filter icons --filter api --filter common build

RUN set -ex; \
    cd /flow-web/; \
    echo "=== Building flow-web ==="; \
    BUILD_PATH=dist pnpm --filter flow-web build; \
    echo "=== Build complete, checking output ==="; \
    ls -la packages/flow-web/

# ========== 第三阶段：运行时镜像 ==========
FROM --platform=${TARGETPLATFORM} ${BASE_IMAGE} AS prod

# 给 nginx 添加权限（flow-web 的配置）
RUN chown root /usr/sbin/nginx && \
    chmod 757 -R /usr/sbin/nginx && \
    chmod u+s /usr/sbin/nginx && \
    chmod 777 -R /var/log/nginx/* && \
    sed -i "s@www-data@root@g" /etc/nginx/nginx.conf && \
    sed -i "s@error_log.*@error_log /dev/null;@g"  /etc/nginx/nginx.conf && \
    sed -i "s@access_log.*@access_log off;@g"  /etc/nginx/nginx.conf && \
    sed -i "s@# server_tokens.*@server_tokens off;@g" /etc/nginx/nginx.conf

# 从构建阶段复制静态文件
COPY --from=vega-builder /vega-web/dist /app/vega-web
COPY --from=flow-builder /flow-web/packages/flow-web/dist /app/flow-web

# 复制 nginx 配置
COPY ./web/docker/docker-multi/nginx-multi.conf /etc/nginx/nginx.conf

# 暴露端口（统一使用 3010）
EXPOSE 3010

# 启动 nginx（不需要指定配置文件，使用默认路径）
CMD ["nginx", "-g", "daemon off;"]
