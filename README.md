# j8s

A Java 8 Stack trace tool for Kubernetes. Interactive CLI for running JVM thread dumps on Java processes in Kubernetes pods.

## Features

- **ASCII logo** displayed on startup
- **Interactive pod selection** from current Kubernetes namespace
- **Container selection** for multi-container pods
- **Automatic jattach deployment** with fallback strategies:
  - Primary: Download locally and copy via `kubectl cp`
  - Fallback: Download directly in container using `curl` or `wget`
- **Java process detection** using `pidof` and `pgrep`
- **Thread dump execution** using jattach
- **No external dependencies** - uses only Go standard library

## Usage

```bash
# Build the application
go build -o j8s main.go

# Or use make
make build

# Show help and logo
./j8s

# Run interactive jstack
./j8s jstack
```

## Prerequisites

- `kubectl` configured and connected to a Kubernetes cluster
- Running Java pods in the current namespace
- Network access for downloading jattach (if not using kubectl cp)

## How it works

1. Lists all running pods in the current namespace using `kubectl get pods -o json`
2. Prompts user to select a pod
3. If multiple containers exist, prompts for container selection
4. Downloads jattach binary and deploys to `/tmp/jattach` in the container
5. Finds Java process PID using `pidof java` or `pgrep -f java`
6. Executes thread dump using `/tmp/jattach <PID> threaddump`
7. Streams output back to the terminal

## Error Handling

The tool provides helpful error messages for common scenarios:
- No kubectl connection
- No pods found in namespace
- No Java processes in selected container
- Failed jattach downloads
- Permission issues

## License

MIT
