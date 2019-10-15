FROM golang:latest

RUN apt-get update && apt-get install -y \
    python2 \
    python3 \
    && rm -rf /var/lib/apt/lists/*

COPY ./ /go/judger

WORKDIR /go/judger

RUN go get -d ./... && go build -o /bin/judger

ENTRYPOINT ["/bin/judger"]
