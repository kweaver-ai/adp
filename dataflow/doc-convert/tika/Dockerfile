ARG BASE_IMAGE=acr.aishu.cn/dip/tika:3.3.0.0-20260325-1
FROM ${BASE_IMAGE}

# 把新增的自定义配置文件复制到镜像内部的执行目录
COPY tika-config.xml /config/tika-config.xml
ENV TIKA_CONFIG="/config/tika-config.xml"

# 显式声明启动命令参数（结合原始的 Entrypoint 一同执行启动服务）
CMD ["-c", "/config/tika-config.xml"]
