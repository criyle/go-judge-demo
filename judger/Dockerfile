FROM golang:latest AS builder

COPY ./ /go/judger/

WORKDIR /go/judger

RUN go get -d  \
    && go build -o /bin/judger

FROM ubuntu:latest

RUN apt-get update && apt-get install -y \
    python2.7 \
    python3 \
    fpc \
    openjdk-11-jdk \
    nodejs \
    golang-go \
    && rm -rf /var/lib/apt/lists/*

COPY --from=builder /bin/judger /bin/judger

WORKDIR /go/judger

ENTRYPOINT ["/bin/judger"]
