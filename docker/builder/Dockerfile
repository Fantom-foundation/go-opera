FROM glider as stage0

# Glider can be found in https://github.com/andrecronje/lachesis/blob/master/docker/glider/Dockerfile

COPY toybox-x86_64 /cp_bin/

RUN mkdir -p "$GOPATH/src/github.com/andrecronje/lachesis" && \
    git clone 'https://github.com/SamuelMarks/lachesis' "$GOPATH/src/github.com/andrecronje/lachesis" && \
    cd "$GOPATH/src/github.com/andrecronje/lachesis" && \
    glide install && \
    cd "$GOPATH/src/github.com/andrecronje/lachesis/cmd/lachesis" && \
    go build -ldflags "-linkmode external -extldflags -static" -a main.go && \
    mv "$GOPATH/src/github.com/andrecronje/lachesis/cmd/lachesis/main" /cp_bin/lachesis

COPY *.c* /build/
COPY *.h* /build/

RUN apk --no-cache add libc-dev && \
    gcc /build/copy.c -o "/cp_bin/copy" -static -static-libgcc -Wno-implicit-function-declaration && \
    gcc /build/env.c  -o "/cp_bin/env"  -static -static-libgcc -Wno-implicit-function-declaration && \
    gcc /build/list.c -o "/cp_bin/list" -static -static-libgcc

RUN apk --no-cache add cmake && \
    git clone https://github.com/SamuelMarks/docker-static-bin /build/docker-static-bin && \
    mkdir /build/docker-static-bin/c/cmake-build-release && \
    cd /build/docker-static-bin/c/cmake-build-release && \
    cmake -DCMAKE_BUILD_TYPE=Release .. && \
    gcc -static ../cmd/crappy_sh.c -I. -Os -o crappy_sh -Wno-implicit-function-declaration -Wno-int-conversion && \
    # strip /docker-static-bin/c/cmake-build-release/crappy_sh && \
    cp /build/docker-static-bin/c/cmake-build-release/crappy_sh /cp_bin/

FROM scratch as lachesis

ENV node_num=0

EXPOSE 1338
EXPOSE 1339
EXPOSE 8000

# cp -r /etc/ssl/certs certs, then add to your `docker build`: `--build-arg ca_certificates=certs`
ARG ca_certificates
ADD "$ca_certificates" /etc/ssl/certs/
COPY --from=0 /cp_bin /bin

# `mkdir -p` cheat
COPY 'Dockerfile' '/lachesis_data_dir/.gitkeep'

COPY peers.json /lachesis_data_dir/
COPY nodes /nodes

# RUN ["/bin/list", "/bin"]
# RUN ["/bin/env"]

ENTRYPOINT ["/bin/crappy_sh", "-v", "-e", "-c", "'/bin/env ; /bin/copy /nodes/$node_num/priv_key.pem /lachesis_data_dir/priv_key.pem ; /bin/list /lachesis_data_dir'"]
# CMD /bin/copy /nodes/$node_num/priv_key.pem /lachesis_data_dir/priv_key.pem
# ENTRYPOINT ["/bin/toybox-x86_64"]
# ENTRYPOINT ["/bin/list", "/lachesis_data_dir"]
# ENTRYPOINT ["/bin/lachesis", "run", "--datadir", "/lachesis_data_dir", "--store_path", "/lachesis_data_dir/badger_db"]
