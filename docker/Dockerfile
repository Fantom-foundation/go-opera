FROM golang:1.12-alpine as builder

RUN apk add --no-cache make gcc musl-dev linux-headers git
                                                                                                                                                                                        
WORKDIR /go/go-lachesis
COPY . .

RUN go mod download

RUN export GIT_COMMIT=$(git rev-list -1 HEAD) && \
    export GIT_DATE=$(git log -1 --date=short --pretty=format:%ct) && \
    export CGO_ENABLED=1 && \
    go build -ldflags "-s -w -X main.gitCommit=$GIT_COMMIT -X main.gitDate=$GIT_DATE" \
    -o /tmp/lachesis \
    ./src/cmd



FROM alpine:latest

RUN apk add --no-cache ca-certificates

COPY --from=builder /tmp/lachesis /

EXPOSE 8545 8546 30303 30303/udp

ENTRYPOINT ["/lachesis"]
