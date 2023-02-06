FROM golang:1.19.4-alpine3.17 as builder
RUN apk add alpine-sdk --no-cache
COPY . /build
WORKDIR /build
RUN go build .
FROM golang:1.19.4-alpine3.17
RUN apk add --no-cache ca-certificates
COPY --from=builder /build/datadog-remote-adapter /bin/datadog-remote-adapter

ENTRYPOINT ["/bin/datadog-remote-adapter"]

