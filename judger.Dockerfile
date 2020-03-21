FROM golang:latest AS builder

WORKDIR /go/judger

COPY go.mod go.sum /go/judger/

RUN go mod download

COPY judger /go/judger/

RUN go build -o /bin/judger

FROM ubuntu:latest

ENV TZ=America/New_York

ENV DEBIAN_FRONTEND=noninteractive

RUN apt-get update && apt-get install -y \
    python2.7 \
    python3 \
    fpc \
    openjdk-11-jdk \
    nodejs \
    golang-go \
    php-cli \
    ghc \
    rustc \
    ruby \
    node-typescript \
    perl6 \
    perl \
    ocaml \
    && rm -rf /var/lib/apt/lists/*

COPY --from=builder /bin/judger /opt/bin/judger

WORKDIR /opt/go/judger

ENTRYPOINT ["/opt/bin/judger"]
