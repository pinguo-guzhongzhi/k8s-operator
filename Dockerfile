FROM mirror.camera360.com/base/golang-builder:1.19.3 as builder
COPY . /app
WORKDIR /app

ENV CGO_ENABLED=0
ENV GO111MODULE=on
#ENV GOPROXY=https://proxy.golang.com.cn,direct
ENV GOPROXY=https://goproxy.io
ENV GOPROXY=https://goproxy.cn
RUN /bin/sh -c 'go build -o app main.go'

# 运维使用的分割线
#---DoNotDelete

#FROM alpine:3.13
FROM mirror.camera360.com/base/centos7.8:basic
WORKDIR /app
COPY --from=builder /app/bin/* /app/bin/
ADD configs /app/configs
EXPOSE 9000/tcp
#ENTRYPOINT [ "/app/bin/app" ]
