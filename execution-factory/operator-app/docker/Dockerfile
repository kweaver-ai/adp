ARG baseImage=acr.aishu.cn/public/ubuntu:22.04.20241202

FROM ${baseImage} AS prod

WORKDIR /app/operator-app

COPY ./operator-app /app/operator-app

# 创建非root用户
RUN useradd -m -r -s /bin/bash aishu \
    && chown -R aishu:aishu /app/operator-app

# 切换到非root用户
USER aishu

CMD ["./operator-app"]