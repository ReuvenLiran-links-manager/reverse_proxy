FROM alpine:latest
RUN apk update && apk upgrade && apk add ca-certificates && rm -rf /var/cache/apk/*
WORKDIR /usr/src/app
COPY bin/reverse-proxy ./
EXPOSE 80
ENTRYPOINT ["./reverse-proxy"]