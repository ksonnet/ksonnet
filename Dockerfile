FROM golang:alpine as build
ARG LD_FLAGS
WORKDIR /go/src/github.com/ksonnet/ksonnet
COPY . /go/src/github.com/ksonnet/ksonnet
RUN CGO_ENABLED=0 GOOS=linux go build -o ks -ldflags="${LD_FLAGS} -s -w" ./cmd/ks

FROM alpine:latest as certs
RUN apk --update add ca-certificates

FROM scratch
COPY --from=certs /etc/ssl/certs/* /etc/ssl/certs/
COPY --from=build /go/src/github.com/ksonnet/ksonnet/ks /bin/ks
VOLUME /tmp

ENTRYPOINT ["/bin/ks"]
