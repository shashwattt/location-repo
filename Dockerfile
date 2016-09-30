FROM golang:1.6.2

ENV GO15VENDOREXPERIMENT=1

RUN apt-get update && apt-get install --no-install-recommends -y \
   ca-certificates curl git-core mercurial \
   g++ dh-autoreconf pkg-config

# Install gflags
RUN apt-get install -y libgflags-dev

# Install snappy
RUN apt-get install -y libsnappy-dev

# Install zlib
RUN apt-get install -y zlib1g-dev

# Install bzip2
RUN apt-get install -y libbz2-dev

# Install Rocksdb
RUN cd /tmp && git clone https://github.com/facebook/rocksdb.git && \
 cd rocksdb && \
 git checkout v4.9 && \
 make shared_lib && \
 mkdir -p /usr/local/rocksdb/lib && \
 mkdir /usr/local/rocksdb/include && \
 cp librocksdb.so* /usr/local/rocksdb/lib && \
 cp /usr/local/rocksdb/lib/librocksdb.so* /usr/lib/ && \
 cp -r include /usr/local/rocksdb/

RUN CGO_CFLAGS="-I/usr/local/rocksdb/include" \
CGO_LDFLAGS="-L/usr/local/rocksdb -lrocksdb -lstdc++ -lm -lz -lbz2 -lsnappy" \
 go get github.com/tecbot/gorocksdb

ADD . /go/src/github.com/locationapi

RUN go get github.com/locationapi

RUN go install github.com/locationapi
 
WORKDIR /go/src/github.com/locationapi

CMD ["/bin/bash", "-c", "go run locationapi.go"]

EXPOSE 8080