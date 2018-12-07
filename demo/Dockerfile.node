FROM offscale/golang-builder-alpine3.8 as stage0

COPY ./ "$GOPATH/src/github.com/Fantom-foundation/go-lachesis/"

RUN mkdir -p /cp_bin /bin \
    && cd "$GOPATH/src/github.com/Fantom-foundation/go-lachesis" \
    && glide install \
    && cd "cmd/lachesis" \
    && go build -ldflags "-linkmode external -extldflags -static -s -w" -a -o /cp_bin/lachesis main.go

RUN apk --no-cache add libc-dev cmake \
    && git clone https://github.com/SamuelMarks/docker-static-bin /build/docker-static-bin \
    && mkdir /build/docker-static-bin/cmake-build-release \
    && cd    /build/docker-static-bin/cmake-build-release \
    && TEST_ENABLED=0 cmake -DCMAKE_BUILD_TYPE=Release .. \
    && cd /build/docker-static-bin/cmd \
    && gcc copy.c      -o "/cp_bin/copy"      -Os -static -Wno-implicit-function-declaration \
    && gcc env.c       -o "/cp_bin/env"       -Os -static -Wno-implicit-function-declaration \
    && gcc list.c      -o "/cp_bin/list"      -Os -static \
    && gcc crappy_sh.c -o "/cp_bin/crappy_sh" -Os -static -Wno-implicit-function-declaration -Wno-int-conversion -I./../cmake-build-release \
    && strip -s /cp_bin/crappy_sh /cp_bin/copy /cp_bin/env /cp_bin/list /cp_bin/lachesis


FROM scratch as lachesis_base

ENV node_addr='127.0.0.1'

EXPOSE 1338
EXPOSE 1339
EXPOSE 8000
EXPOSE 12000

VOLUME /data

# cp -r /etc/ssl/certs certs, then add to your `docker build`: `--build-arg ca_certificates=certs`
ARG ca_certificates
ADD "$ca_certificates" /etc/ssl/certs/
COPY --from=0 /cp_bin /bin

ENTRYPOINT ["/bin/lachesis"]

#CMD [ "/bin/crappy_sh", "-v", "-e", "-c", "/bin/env ; /bin/lachesis run --datadir /data --store /data/badger_db --listen=$node_addr:12000 --heartbeat=50s" ]
