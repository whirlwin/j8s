# j8s

[![Build and Publish](https://github.com/whirlwin/j8s/actions/workflows/build-and-publish.yaml/badge.svg?branch=v0.1.0)](https://github.com/whirlwin/j8s/actions/workflows/build-and-publish.yaml)

A Java 8 Stack trace debug for Kubernetes. Interactive CLI for running JVM thread dumps and heap dumps on Java processes in Kubernetes Pods where there is no native jmap or jstack tool available.


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

## Prerequisites

- `kubectl` configured and connected to a Kubernetes cluster
- Running Java pods in the current namespace
- Network access for downloading jattach (if not using kubectl cp)

## Installation
Download any of the [releases](https://github.com/whirlwin/j8s/releases), uncompress, and add to your executable PATH.