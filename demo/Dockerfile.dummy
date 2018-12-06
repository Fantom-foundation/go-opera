FROM offscale/golang-builder-alpine3.8 as stage0

COPY ./ "$GOPATH/src/github.com/Fantom-foundation/go-lachesis/"

RUN mkdir -p /cp_bin /bin \
    && cd "$GOPATH/src/github.com/Fantom-foundation/go-lachesis" \
    && glide install \
    && cd "cmd/dummy" \
    && go build -ldflags "-linkmode external -extldflags -static -s -w" -a -o /cp_bin/dummy main.go

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
    && strip -s /cp_bin/crappy_sh /cp_bin/copy /cp_bin/env /cp_bin/list /cp_bin/dummy


FROM scratch as lachesis_base

# cp -r /etc/ssl/certs certs, then add to your `docker build`: `--build-arg ca_certificates=certs`
ARG ca_certificates
ADD "$ca_certificates" /etc/ssl/certs/
COPY --from=0 /cp_bin /bin

ENTRYPOINT ["/bin/dummy"]
