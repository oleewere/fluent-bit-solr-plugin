FROM golang:1.11-alpine
RUN apk add --no-cache gcc musl-dev
ADD . /go/src/github.com/oleewere/fluent-bit-solr-plugin
WORKDIR /go/src/github.com/oleewere/fluent-bit-solr-plugin
RUN go build -buildmode=c-shared -ldflags="-s -w" -o out_solr.so out_solr.go

FROM fluent/fluent-bit:1.0.4
COPY --from=0 /go/src/github.com/oleewere/fluent-bit-solr-plugin/out_solr.so /fluent-bit/out_solr.so
COPY --from=0 /lib/libc.musl-x86_64.so.* /lib/

CMD ["/fluent-bit/bin/fluent-bit", "-e", "/fluent-bit/out_solr.so", "-c", "/fluent-bit/etc/fluent-bit.conf"]
