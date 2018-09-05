FROM glider as stage0

# Glider can be found in https://github.com/andrecronje/lachesis/blob/master/docker/glider/Dockerfile

RUN mkdir -p "$GOPATH/src/github.com/andrecronje/lachesis" && \
    git clone 'https://github.com/andrecronje/lachesis' "$GOPATH/src/github.com/andrecronje/lachesis" && \
    cd "$GOPATH/src/github.com/andrecronje/lachesis" && \
    glide install && \
    cd "$GOPATH/src/github.com/andrecronje/lachesis/cmd/lachesis" && \
    go build -ldflags "-linkmode external -extldflags -static" -a main.go && \
    mv "$GOPATH/src/github.com/andrecronje/lachesis/cmd/lachesis/main" "$GOPATH/bin/lachesis"

FROM scratch as lachesis

ENV node_num=0

EXPOSE 1338
EXPOSE 1339
EXPOSE 8000

# cp -r /etc/ssl/certs certs, then add to your `docker build`: `--build-arg ca_certificates=certs`
ARG ca_certificates
ADD "$ca_certificates" /etc/ssl/certs/
COPY --from=0 /go/bin/lachesis /lachesis

# `mkdir -p` cheat
COPY 'Dockerfile' "/lachesis_data_dir/.gitkeep"

COPY peers.json /lachesis_data_dir/
COPY nodes '/nodes'
COPY "nodes/$node_num/priv_key.pem" '/lachesis_data_dir/priv_key.pem'

ENTRYPOINT ["/lachesis", "run", "--datadir", "/lachesis_data_dir", "--store_path", "/lachesis_data_dir/badger_db"]
