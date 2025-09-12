# j8s

[![Build and Publish](https://github.com/whirlwin/j8s/actions/workflows/build-and-publish.yaml/badge.svg)](https://github.com/whirlwin/j8s/actions/workflows/build-and-publish.yaml)

A Java 8 Stack trace debug for Kubernetes. Interactive CLI for running JVM thread dumps and heap dumps on Java processes in Kubernetes Pods.


```text

  ▖▞▀▖
 ▗▖▚▄▘▞▀▘
  ▌▌ ▌▝▀▖
 ▄▘▝▀ ▀▀

j8s is a Java CLI tool for Kubernetes

Usage:
  j8s dump threads    Interactive JVM thread dump from Kubernetes pod
  j8s dump heap       Interactive JVM heap dump from Kubernetes pod (downloads locally)
  j8s                 Show this help
```

## Usage

```bash
# Build the application
go build -o j8s main.go

# Or use make
make build

# Show help and logo
./j8s

# Run interactive thread dump
./j8s dump threads

# Run interactive heap dump (downloads .hprof file locally)
./j8s dump heap
```

## Prerequisites

- `kubectl` configured and connected to a Kubernetes cluster
- Running Java pods in the current namespace
- Network access for downloading jattach (if not using kubectl cp)

