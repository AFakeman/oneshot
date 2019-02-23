FROM golang:1.11

WORKDIR /go/src/github.com/afakeman/oneshot/
COPY . .
RUN go get -v
RUN go install -v

CMD ["oneshot"]
