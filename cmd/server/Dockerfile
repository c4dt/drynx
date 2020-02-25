FROM alpine:latest

RUN apk add dash
COPY ./server-static /server

ENTRYPOINT ["/server"]
