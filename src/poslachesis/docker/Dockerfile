FROM golang:1.12 as build

WORKDIR /go/src/github.com/Fantom-foundation/go-lachesis
COPY . .

RUN go get github.com/Masterminds/glide 

RUN glide install && CGO_ENABLED=0 go build -ldflags "-s -w" -o /tmp/lachesis ./src/poslachesis/cli



FROM scratch as prod

COPY --from=build /tmp/lachesis /

EXPOSE 55555 55556 55557

ENTRYPOINT ["/lachesis"]
