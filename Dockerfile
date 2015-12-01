FROM golang:1.5
MAINTAINER Hugo González Labrador

ENV CLAWIO_BENCH_AUTH_ADDR service-auth:57000
ENV CLAWIO_BENCH_META_ADDR service-localstore-meta:57001
ENV CLAWIO_BENCH_DATA_ADDR http://service-localstore-data:57002

ADD . /go/src/github.com/clawio/clawiobench
WORKDIR /go/src/github.com/clawio/clawiobench

RUN go get -u github.com/tools/godep
RUN godep restore
RUN go install
