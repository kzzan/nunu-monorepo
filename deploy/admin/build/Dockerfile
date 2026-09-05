ARG REGISTRY=docker.io
FROM ${REGISTRY}/golang:1.26-alpine AS builder
RUN set -eux && sed -i 's/dl-cdn.alpinelinux.org/mirrors.ustc.edu.cn/g' /etc/apk/repositories
RUN apk add --no-cache nodejs npm

ARG APP_RELATIVE_PATH
ARG BUILD_ADMIN_WEB=true

COPY . /data/app
WORKDIR /data/app

RUN if [ "${BUILD_ADMIN_WEB}" = "true" ]; then npm install -g pnpm && HUSKY=0 pnpm --dir /data/app/app/admin/web install --frozen-lockfile && pnpm --dir /data/app/app/admin/web build; fi
RUN rm -rf /data/app/bin/
RUN export GOPROXY=https://goproxy.cn,direct && go mod tidy && go build -ldflags="-s -w" -o ./bin/server ${APP_RELATIVE_PATH}
RUN mv config /data/app/bin/


FROM ${REGISTRY}/alpine:3.16
RUN set -eux && sed -i 's/dl-cdn.alpinelinux.org/mirrors.ustc.edu.cn/g' /etc/apk/repositories


RUN apk add tzdata && cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime \
    && echo "Asia/Shanghai" > /etc/timezone \
    && apk del tzdata


ARG APP_ENV
ENV APP_ENV=${APP_ENV}
ARG APP_CONF=config/admin/prod.yml
ENV APP_CONF=${APP_CONF}

WORKDIR /data/app
COPY --from=builder /data/app/bin /data/app

EXPOSE 8000
ENTRYPOINT [ "./server" ]

#docker build -t 1.1.1.1:5000/demo-api:v1 --build-arg APP_CONF=config/admin/prod.yml --build-arg APP_RELATIVE_PATH=./app/admin/cmd/server .
#docker run -it --rm --entrypoint=ash 1.1.1.1:5000/demo-api:v1
