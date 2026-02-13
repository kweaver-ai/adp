# Copyright The kweaver.ai Authors.
# Licensed under the Apache License, Version 2.0.
# See the LICENSE file in the project root for details.

FROM acr.aishu.cn/af/openjdk:8u342-bullseye-all
ENV TZ=Asia/Shanghai
ENV JAVA_OPTS="-XX:+UseContainerSupport -XX:InitialRAMPercentage=70.0 -XX:MaxRAMPercentage=70.0 -XX:MaxMetaspaceSize=512M -XX:MetaspaceSize=128M -XX:+UseG1GC -XX:InitiatingHeapOccupancyPercent=30 -XX:ParallelGCThreads=8 -XX:+ExitOnOutOfMemoryError"

# Presto version will be passed in at build time
ARG PRESTO_VERSION

RUN ln -sf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone && \
    sed -i 's/http:\/\/deb.debian.org\/debian/http:\/\/mirrors.aliyun.com\/debian/g' /etc/apt/sources.list && \
    sed -i 's/http:\/\/security.debian.org\/debian-security/http:\/\/mirrors.aliyun.com\/debian-security/g' /etc/apt/sources.list && \
    mkdir -p /opt/vega && \
    mkdir -p /opt/vega/config

WORKDIR /opt/vega
ADD ./data-connection.jar ./data-connection.jar

# Specify the entrypoint to start
CMD java -Dfile.encoding=UTF-8 $JAVA_OPTS -jar -server -agentlib:jdwp=transport=dt_socket,server=y,suspend=n,address=5005 data-connection.jar --spring.config.location=./config/
