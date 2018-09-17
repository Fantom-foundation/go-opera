FROM glider as stage0

# Glider can be found in https://github.com/andrecronje/lachesis/blob/master/docker/glider/Dockerfile

COPY toybox-x86_64 /cp_bin/

RUN mkdir -p "$GOPATH/src/github.com/andrecronje/lachesis" && \
    git clone 'https://github.com/andrecronje/lachesis' "$GOPATH/src/github.com/andrecronje/lachesis" && \
    cd "$GOPATH/src/github.com/andrecronje/lachesis" && \
    glide install && \
    cd "$GOPATH/src/github.com/andrecronje/lachesis/cmd/lachesis" && \
    go build -ldflags "-linkmode external -extldflags -static" -a main.go && \
    mv "$GOPATH/src/github.com/andrecronje/lachesis/cmd/lachesis/main" /cp_bin/lachesis

COPY *.c* /build/
COPY *.h* /build/

RUN apk --no-cache add libc-dev && \
    gcc /build/copy.c -static -o "/cp_bin/copy" -Wno-implicit-function-declaration -static-libgcc && \
    gcc /build/env.c -static -o "/cp_bin/env" -Wno-implicit-function-declaration && \
    gcc /build/list.c -static -o "/cp_bin/list"

RUN git clone https://github.com/SamuelMarks/docker-static-bin /docker-static-bin && \
    mkdir /docker-static-bin/c/cmake-build-release && \
    cd /docker-static-bin/c/cmake-build-release \
    cmake -DCMAKE_BUILD_TYPE=Release .. && \
    ls -a # && \
#    gcc ../cmd/crappy_sh.c -I. -Os -o crappy_sh && \
#    strip /docker-static-bin/c/cmake-build-release/crappy_sh && cp /docker-static-bin/c/cmake-build-release/crappy_sh /cp_bin/

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
COPY 'Dockerfile' "/lachesis_data_dir/.gitkeep"

COPY peers.json /lachesis_data_dir/
COPY nodes /nodes

CMD ["/bin/crappy_sh", "-e", "-c", "'/bin/copy /nodes/$node_num/priv_key.pem /lachesis_data_dir/priv_key.pem'"]
# CMD /bin/copy /nodes/$node_num/priv_key.pem /lachesis_data_dir/priv_key.pem
# ENTRYPOINT ["/bin/toybox-x86_64"]
ENTRYPOINT ["/bin/list", "/lachesis_data_dir"]
# ENTRYPOINT ["/bin/lachesis", "run", "--datadir", "/lachesis_data_dir", "--store_path", "/lachesis_data_dir/badger_db"]
