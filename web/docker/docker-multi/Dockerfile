ARG BASE_IMAGE=acr.aishu.cn/public/nginx:1.22.1-alpine

# ========== 运行时镜像 ==========
FROM ${BASE_IMAGE}

# 给 nginx 添加权限
RUN chown root /usr/sbin/nginx && \
    chmod 757 -R /usr/sbin/nginx && \
    chmod u+s /usr/sbin/nginx && \
    chmod 777 -R /var/log/nginx/* && \
    sed -i "s@www-data@root@g" /etc/nginx/nginx.conf && \
    sed -i "s@error_log.*@error_log /dev/null;@g"  /etc/nginx/nginx.conf && \
    sed -i "s@access_log.*@access_log off;@g"  /etc/nginx/nginx.conf && \
    sed -i "s@# server_tokens.*@server_tokens off;@g" /etc/nginx/nginx.conf

# 复制静态文件（已编译好的产物）
COPY vega-web /app/vega-web
COPY flow-web /app/flow-web

# 复制 nginx 配置
COPY nginx-multi.conf /etc/nginx/nginx.conf

EXPOSE 3010

CMD ["nginx", "-g", "daemon off;"]
