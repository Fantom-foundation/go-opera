FROM glider as stage0

# Glider can be found in https://github.com/andrecronje/lachesis/blob/master/docker/glider/Dockerfile

RUN mkdir -p "$GOPATH/src/github.com/andrecronje/lachesis" /cp_bin && \
    git clone 'https://github.com/SamuelMarks/lachesis' "$GOPATH/src/github.com/andrecronje/lachesis" && \
    cd "$GOPATH/src/github.com/andrecronje/lachesis" && \
    glide install && \
    cd "$GOPATH/src/github.com/andrecronje/lachesis/cmd/lachesis" && \
    go build -ldflags "-linkmode external -extldflags -static" -a main.go && \
    mv "$GOPATH/src/github.com/andrecronje/lachesis/cmd/lachesis/main" /cp_bin/lachesis

RUN apk --no-cache add libc-dev cmake && \
    git clone https://github.com/SamuelMarks/docker-static-bin /build/docker-static-bin && \
    mkdir /build/docker-static-bin/c/cmake-build-release && \
    cd    /build/docker-static-bin/c/cmake-build-release && \
    cmake -DCMAKE_BUILD_TYPE=Release .. && \
    cd /build/docker-static-bin/c/cmd && \
    gcc copy.c      -o "/cp_bin/copy"      -Os -static -static-libgcc -Wno-implicit-function-declaration && \
    gcc env.c       -o "/cp_bin/env"       -Os -static -static-libgcc -Wno-implicit-function-declaration && \
    gcc list.c      -o "/cp_bin/list"      -Os -static -static-libgcc && \
    gcc crappy_sh.c -o "/cp_bin/crappy_sh" -Os -static -Wno-implicit-function-declaration -Wno-int-conversion -I./../cmake-build-release && \
    strip /cp_bin/crappy_sh

FROM scratch as lachesis

ENV node_num=0
ENV node_addr='127.0.0.1'

EXPOSE 1338
EXPOSE 1339
EXPOSE 8000
EXPOSE 12000

# cp -r /etc/ssl/certs certs, then add to your `docker build`: `--build-arg ca_certificates=certs`
ARG ca_certificates
ADD "$ca_certificates" /etc/ssl/certs/
COPY --from=0 /cp_bin /bin

COPY peers.json /lachesis_data_dir/
COPY nodes /nodes

ENTRYPOINT ["/bin/crappy_sh", "-v", "-e", "-c", "/bin/env ; /bin/list /cp_bin ;  /bin/copy /nodes/$node_num/priv_key.pem /lachesis_data_dir/priv_key.pem ; /bin/list /lachesis_data_dir ; /bin/lachesis run --datadir /lachesis_data_dir --store_path /lachesis_data_dir/badger_db -node_addr=$node_addr:12000 -heartbeat=100 -no_client"]
