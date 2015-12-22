FROM golang:1.5
MAINTAINER Hugo González Labrador

ENV CLAWIO_BENCH_AUTH_ADDR service-auth:57000
ENV CLAWIO_BENCH_META_ADDR service-localfs-meta:57001
ENV CLAWIO_BENCH_DATA_ADDR http://service-localfs-data:57002

ADD . /go/src/github.com/clawio/clawiobench
WORKDIR /go/src/github.com/clawio/clawiobench
RUN go get code.google.com/p/go-uuid/uuid \
    github.com/Sirupsen/logrus \
    github.com/cheggaaa/pb \
    github.com/golang/protobuf/proto \
    github.com/spf13/cobra \
    github.com/spf13/viper \
    golang.org/x/net/context \
    google.golang.org/grpc \ 
    google.golang.org/grpc/codes

RUN go install

CMD ["clawiobench"]
