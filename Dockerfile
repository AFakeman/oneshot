FROM golang:1.11 as builder

WORKDIR /go/src/github.com/afakeman/oneshot/
COPY . .
RUN go get -v
RUN go install -v

FROM alpine

RUN apk add --no-cache \
        libc6-compat
COPY --from=builder /go/bin/oneshot /usr/bin/oneshot

CMD ["oneshot"]
