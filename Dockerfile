FROM golang:1.13 AS builder
WORKDIR /opt/observability-demo
COPY . .

RUN go vet
RUN CGO_ENABLED=0 go build

FROM alpine:latest
RUN apk update && apk add ca-certificates && rm -rf /var/cache/apk/*

WORKDIR /opt/bin
COPY --from=builder /opt/observability-demo/observability-demo ./

CMD ./observability-demo
