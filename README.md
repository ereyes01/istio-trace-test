# istio-trace-test
A test of trace with GKE + Istio + OpenCensus

## Prerequisites

To build, you need to have a GKE cluster handy, with Istio enabled on it. [Get it running](https://cloud.google.com/istio/docs/istio-on-gke/installing).

You also need to be authenticated / connected to your GKE cluster so that `kubectl` commands work.

## Deploy

Spin up Istio config and Kubernetes pods:

```
kubectl apply -f istio.yml
kubectl apply -f k8s.yml
```

## Build locally

You'll need a Go version that supports Go modules. Just `go build` will do.
