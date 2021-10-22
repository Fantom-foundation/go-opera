FROM golang:1.14-alpine as builder

RUN apk add --no-cache make gcc musl-dev linux-headers

ADD . /fantom
RUN cd /fantom && make all

FROM alpine:latest

RUN apk add --no-cache ca-certificates
COPY --from=builder /fantom/build/opera /usr/local/bin/

EXPOSE 5050 18545 18546 18547 19090

ENTRYPOINT ["opera"]
