FROM golang:1.13 AS builder
WORKDIR /opt/observability-demo

# speed up the build by allowing docker to cache deps
ENV GO111MODULE=on
COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .

RUN go vet
RUN CGO_ENABLED=0 go build

FROM alpine:latest
RUN apk update && apk add ca-certificates && rm -rf /var/cache/apk/*

WORKDIR /opt/bin
COPY --from=builder /opt/observability-demo/istio-trace-test ./

CMD ./istio-trace-test
