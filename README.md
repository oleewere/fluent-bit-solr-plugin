# Fluent-Bit Solr Output Plugin

[![Docker Pulls](https://img.shields.io/docker/pulls/oleewere/fluent-bit.svg)](https://hub.docker.com/r/oleewere/fluent-bit/)
[![Go Report Card](https://goreportcard.com/badge/github.com/oleewere/fluent-bit-solr-plugin)](https://goreportcard.com/report/github.com/oleewere/fluent-bit-solr-plugin)
![license](http://img.shields.io/badge/license-Apache%20v2-blue.svg)

Fluent Bit Solr plugin

#### Properties

- `Url`: url of the Solr host (or proxy)
- `Collection`: Solr collection
- `Context`: Context (uri) for the URL. Default `/solr`
- `TimeSolrField`: Field name that will be generated for the documents as timestamp. (Default format: `2006-01-02T15:04:05.000`)
- `Epoch`: Use unix epoch for the `TimeSolrField`. Default: `false`

#### Usage

Build & Run with fluent-bit:

```bash
go build -buildmode=c-shared -ldflags="-s -w" -o out_solr.so out_solr.go
/fluent-bit/bin/fluent-bit -e /fluent-bit/out_solr.so -c /fluent-bit/etc/fluent-bit.conf
```

Use with docker:

```
# /fluent-bit/etc/fluent-bit.conf can be used from a volume
docker run --rm oleewere/fluent-bit
```



#### Example configuration

```ini
[SERVICE]
    Daemon       Off
    Log_Level    debug

[INPUT]
    Name          tail
    Path          examples/example.log
    Parser        myparser

[OUTPUT]
    Name       solr
    Match      *
    Collection hadoop_logs
    Url        http://localhost:8983
