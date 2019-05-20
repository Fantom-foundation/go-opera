FROM offscale/golang-builder-alpine3.8 as stage0

RUN export GOPATH=$(go env GOPATH) && \
    mkdir -p "$GOPATH/src/github.com/Fantom-foundation/go-lachesis" /cp_bin /bin && \
    git clone 'https://github.com/SamuelMarks/lachesis' "$GOPATH/src/github.com/Fantom-foundation/go-lachesis" && \
    cd "$GOPATH/src/github.com/Fantom-foundation/go-lachesis" && \
    glide install && \
    cd "$GOPATH/src/github.com/Fantom-foundation/go-lachesis/cmd/lachesis" && \
    go build -ldflags "-linkmode external -extldflags -static -s -w" -a main.go && \
    mv "$GOPATH/src/github.com/Fantom-foundation/go-lachesis/cmd/lachesis/main" /cp_bin/lachesis

# ADD https://github.com/upx/upx/releases/download/v3.95/upx-3.95-amd64_linux.tar.xz /tmp
# RUN /bin/tar --version && \
#     sha512sum /tmp/upx-3.95-amd64_linux.tar.xz && \
#     /bin/tar xf /tmp/upx-3.95-amd64_linux.tar.xz && \
#     /bin/tar -C /bin --strip-components=1 -xzf /tmp/upx-3.95-amd64_linux.tar.xz

# COPY upx /cp_bin/

RUN apk --no-cache add libc-dev cmake && \
    git clone https://github.com/SamuelMarks/docker-static-bin /build/docker-static-bin && \
    mkdir /build/docker-static-bin/cmake-build-release && \
    cd    /build/docker-static-bin/cmake-build-release && \
    TEST_ENABLED=0 cmake -DCMAKE_BUILD_TYPE=Release .. && \
    cd /build/docker-static-bin/cmd && \
    gcc copy.c      -o "/cp_bin/copy"      -Os -static -Wno-implicit-function-declaration && \
    gcc env.c       -o "/cp_bin/env"       -Os -static -Wno-implicit-function-declaration && \
    gcc list.c      -o "/cp_bin/list"      -Os -static && \
    gcc crappy_sh.c -o "/cp_bin/crappy_sh" -Os -static -Wno-implicit-function-declaration -Wno-int-conversion -I./../cmake-build-release && \
    strip -s /cp_bin/crappy_sh /cp_bin/copy /cp_bin/env /cp_bin/list /cp_bin/lachesis
    # /cp_bin/upx --brute /cp_bin/lachesis /cp_bin/crappy_sh /cp_bin/copy /cp_bin/list

FROM scratch as lachesis_base

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

# /cp_bin/upx -d /cp_bin/lachesis /cp_bin/crappy_sh /cp_bin/copy /cp_bin/list ;
ENTRYPOINT ["/bin/crappy_sh", "-v", "-e", "-c", "/bin/env ; /bin/list /cp_bin ; /bin/copy /nodes/$node_num/priv_key.pem /lachesis_data_dir/priv_key.pem ; /bin/list /lachesis_data_dir ; /bin/lachesis run /bin/lachesis run --datadir /lachesis_data_dir --store /lachesis_data_dir/badger_db --listen=$node_addr:12000 --heartbeat=100s"]
