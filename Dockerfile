FROM golang:1.5
MAINTAINER Hugo González Labrador

ENV CLAWIO_BENCH_AUTH_ADDR service-auth:57000
ENV CLAWIO_BENCH_META_ADDR service-localstore-meta:57001
ENV CLAWIO_BENCH_DATA_ADDR http://service-localstore-data:57002

ADD . /go/src/github.com/clawio/clawiobench
WORKDIR /go/src/github.com/clawio/clawiobench

RUN go get -u github.com/tools/godep
RUN godep restore

RUN go get code.google.com/p/go-uuid/uuid
RUN go get github.com/Sirupsen/logrus
RUN go get github.com/cheggaaa/pb
RUN go get github.com/golang/protobuf/proto
RUN go get github.com/spf13/cobra
RUN go get github.com/spf13/viper
RUN go get golang.org/x/net/context
RUN go get google.golang.org/grpc
RUN go get google.golang.org/grpc/codes

RUN go install
