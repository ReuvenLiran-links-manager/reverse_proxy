FROM alpine:latest
WORKDIR /usr/src/app
COPY bin/reverse-proxy ./
EXPOSE 80
ENTRYPOINT ["./reverse-proxy"]