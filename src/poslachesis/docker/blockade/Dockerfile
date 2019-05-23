FROM golang:1.12

WORKDIR /go/src/github.com/Fantom-foundation/go-lachesis
COPY . .

RUN go get github.com/Masterminds/glide 

RUN glide install && CGO_ENABLED=0 go build -ldflags "-s -w" -o /tmp/lachesis ./src/poslachesis/cli

EXPOSE 55555 55556 55557

ENTRYPOINT ["/tmp/lachesis"]
