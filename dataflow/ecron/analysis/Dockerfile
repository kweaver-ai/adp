ARG BASE_IMAGE=acr.aishu.cn/public/ubuntu:22.04.20251014
FROM ${BASE_IMAGE}

ADD /bin/ecron-analysis /ecron/ecron-analysis

WORKDIR /ecron/

CMD [ "./ecron-analysis" ]
