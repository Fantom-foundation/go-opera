FROM glider as stage0

# Glider can be found in https://github.com/andrecronje/evm/blob/master/docker/glider/Dockerfile

RUN mkdir -p "$GOPATH/src/github.com/andrecronje/lachesis" /cp_bin /bin
COPY . "$GOPATH/src/github.com/andrecronje/lachesis"
RUN cd "$GOPATH/src/github.com/andrecronje/lachesis" && \
    rm -rf vendor && \
    glide install && \
    cd "$GOPATH/src/github.com/andrecronje/lachesis/cmd/lachesis" && \
    go build -ldflags "-linkmode external -extldflags -static -s -w" -a main.go && \
    mv "$GOPATH/src/github.com/andrecronje/lachesis/cmd/lachesis/main" /cp_bin/lachesis

# ADD https://github.com/upx/upx/releases/download/v3.95/upx-3.95-amd64_linux.tar.xz /tmp
# RUN /bin/tar --version && \
#     md5sum /tmp/upx-3.95-amd64_linux.tar.xz && \
#     /bin/tar xf /tmp/upx-3.95-amd64_linux.tar.xz && \
#     /bin/tar -C /bin --strip-components=1 -xzf /tmp/upx-3.95-amd64_linux.tar.xz

ARG compress=false

COPY docker/builder/upx /cp_bin/

RUN apk --no-cache add libc-dev cmake && \
    git clone https://github.com/SamuelMarks/docker-static-bin /build/docker-static-bin && \
    mkdir /build/docker-static-bin/c/cmake-build-release && \
    cd    /build/docker-static-bin/c/cmake-build-release && \
    cmake -DCMAKE_BUILD_TYPE=Release .. && \
    cd /build/docker-static-bin/c/cmd && \
    gcc copy.c      -o "/cp_bin/copy"      -Os -static -Wno-implicit-function-declaration && \
    gcc env.c       -o "/cp_bin/env"       -Os -static -Wno-implicit-function-declaration && \
    gcc list.c      -o "/cp_bin/list"      -Os -static && \
    gcc crappy_sh.c -o "/cp_bin/crappy_sh" -Os -static -Wno-implicit-function-declaration -Wno-int-conversion -I./../cmake-build-release
    # $compress && \
    # strip -s /cp_bin/crappy_sh /cp_bin/copy /cp_bin/env /cp_bin/list /cp_bin/lachesis && \
    # /cp_bin/upx --brute /cp_bin/lachesis /cp_bin/crappy_sh /cp_bin/copy /cp_bin/list

FROM scratch as lachesis_base

EXPOSE 1338
EXPOSE 1339
EXPOSE 8000
EXPOSE 9000
EXPOSE 12000

# cp -r /etc/ssl/certs certs, then add to your `docker build`: `--build-arg ca_certificates=certs`
ARG ca_certificates=certs
COPY "$ca_certificates" /etc/ssl/certs/

ENV node_num=0
ENV node_addr='127.0.0.1'

COPY --from=0 /cp_bin /bin

COPY peers.json /lachesis_data_dir/
COPY nodes /nodes

# /cp_bin/upx -d /cp_bin/lachesis /cp_bin/crappy_sh /cp_bin/copy /cp_bin/list ;
ENTRYPOINT ["/bin/crappy_sh", "-v", "-e", "-c", "/bin/env ; /bin/list /cp_bin ; /bin/copy /nodes/$node_num/priv_key.pem /lachesis_data_dir/priv_key.pem ; /bin/list /lachesis_data_dir ; /bin/lachesis run --test --datadir /lachesis_data_dir --store_path /lachesis_data_dir/badger_db -node_addr=$node_addr:12000 -proxy_addr=$node_addr:9000 -heartbeat=100 -no_client"]
