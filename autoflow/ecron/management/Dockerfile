ARG BASE_IMAGE=acr.aishu.cn/public/ubuntu:22.04.20251014
FROM ${BASE_IMAGE}

ADD /bin/ecron-management /ecron/ecron-management

WORKDIR /ecron/

CMD [ "./ecron-management" ]
