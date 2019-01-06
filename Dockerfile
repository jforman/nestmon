FROM golang:1.11 as builder
WORKDIR /go/src/github.com/jforman/nestmon/
RUN go get -d -v github.com/influxdata/influxdb/client/v2
COPY nestmon.go stream.go structs.go ./
COPY examples/thermostat_status.go examples/thermostat_streaming_status.go bin/
RUN CGO_ENABLED=0 go build -a -installsuffix cgo -o nestmon_poller bin/thermostat_status.go
RUN CGO_ENABLED=0 go build -a -installsuffix cgo -o nestmon_streaming_status bin/thermostat_streaming_status.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /
COPY --from=builder /go/src/github.com/jforman/nestmon/nestmon_poller .
COPY --from=builder /go/src/github.com/jforman/nestmon/nestmon_streaming_status .
