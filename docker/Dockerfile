FROM fedora:34
RUN dnf install -y git make gcc gcc-c++ which iproute iputils procps-ng vim-minimal tmux net-tools htop tar jq npm openssl-devel perl rust cargo golang

ADD https://github.com/Fantom-foundation/go-opera/archive/refs/tags/v1.1.0-rc.5.tar.gz /go-opera/
RUN echo BEFORE && go env | grep GOPROXY
ENV GOPROXY=
RUN echo AFTER && go env | grep GOPROXY
RUN cd /go-opera && tar -xvf * && cd ./go-opera-1.1.0-rc.5 && make opera && mv ./build/opera /usr/bin/opera
