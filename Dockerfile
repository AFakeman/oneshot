FROM docker:18.09.2 as docker
FROM golang:1.11 as builder

WORKDIR /go/src/github.com/afakeman/oneshot/
COPY . .
RUN go get -v
RUN go install -v

FROM alpine

RUN apk add --no-cache \
        libc6-compat
COPY --from=builder /go/bin/oneshot /usr/bin/oneshot
COPY --from=docker /usr/local/bin/docker /usr/bin/docker

CMD ["oneshot"]
