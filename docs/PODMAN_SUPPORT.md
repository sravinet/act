# Podman Support in Act

Act now supports both Docker and Podman container runtimes, with automatic detection and seamless switching between them.

## Features

### ✅ Automatic Runtime Detection

Act automatically detects and uses the best available container runtime:

1. **Podman preferred** (better security, rootless, no daemon)
2. **Docker fallback** (broader compatibility)
3. **User override** (via CLI flags or environment variables)

### ✅ CLI Configuration

```bash
# Automatic detection (default)
act

# Force specific runtime
act --container-runtime=podman
act --container-runtime=docker

# Custom socket path
act --container-socket=/custom/path/to/socket

# Combined options
act --container-runtime=podman --container-socket=/run/user/1000/podman/podman.sock
```

### ✅ Environment Variables

```bash
# Set preferred runtime
export ACT_CONTAINER_RUNTIME=podman
act

# Set custom socket
export ACT_CONTAINER_SOCKET=/custom/socket
act
```

### ✅ Zero Breaking Changes

All existing Docker workflows continue to work unchanged. Podman support is additive and backward-compatible.

## Architecture

### Runtime Detection Priority

1. **CLI flags**: `--container-runtime` and `--container-socket`
2. **Environment variables**: `ACT_CONTAINER_RUNTIME`, `ACT_CONTAINER_SOCKET`
3. **Runtime-specific env vars**: `PODMAN_HOST`, `DOCKER_HOST`
4. **Auto-detection**: Socket availability + binary verification

### Implementation Strategy

- **Docker-compatible API**: Podman's Docker-compatible REST API for maximum compatibility
- **Runtime-specific optimizations**: Podman-specific handling for rootless containers, networking, and error messages
- **Graceful fallback**: Intelligent error handling and user guidance

## Podman Advantages

When Podman is available, you get:

- **🔒 Enhanced Security**: Rootless containers by default
- **🚫 No Daemon**: Podman doesn't require a background daemon
- **⚡ Better Performance**: Lower resource overhead
- **🐧 Linux Native**: Better integration with systemd and Linux security features

## Quick Start

### Install Podman

```bash
# macOS
brew install podman
podman machine init
podman machine start

# Ubuntu/Debian
sudo apt-get install podman

# Fedora/RHEL/CentOS
sudo dnf install podman
```

### Use with Act

```bash
# Act will automatically detect and use Podman
act

# Verify Podman is being used (look for "podman" in log messages)
act --verbose

# Force Podman if you have both runtimes
act --container-runtime=podman
```

## Troubleshooting

### Check Available Runtimes

```bash
# Check if Podman is detected
act --container-runtime=podman --dryrun -l

# Check if Docker is detected  
act --container-runtime=docker --dryrun -l
```

### Common Issues

1. **"No container runtime detected"**
   - Install Docker or Podman
   - Ensure the daemon/socket is running
   - Check socket permissions

2. **Podman not detected**
   - Verify Podman installation: `podman version`
   - Check socket location: `podman info --format=json | jq .host.remoteSocket`
   - Start Podman service if needed: `systemctl --user start podman.socket`

3. **Permission denied**
   - For Docker: Add user to docker group or run with sudo
   - For Podman: Use rootless mode (default) or check socket permissions

### Debug Information

```bash
# Enable debug logging
act --verbose --container-runtime=auto

# Test specific runtime
act --container-runtime=podman --dryrun

# Use custom socket
act --container-socket=unix:///run/user/$(id -u)/podman/podman.sock
```

## Technical Details

For implementation details and architectural decisions, see:

- [ADR-001: Container Runtime Abstraction](./adr/ADR-001-container-runtime-abstraction.md)
- [ADR-002: Podman Integration Strategy](./adr/ADR-002-podman-integration-strategy.md)
- [ADR-003: Runtime Detection Mechanism](./adr/ADR-003-runtime-detection-mechanism.md)

## Migration Notes

### From Docker-only Act

No changes required! Your existing workflows and configurations continue to work exactly as before.

### Optimizing for Podman

If you want to take advantage of Podman-specific features:

1. Use `--container-runtime=podman` to ensure Podman is used
2. Consider rootless container workflows for better security
3. Leverage systemd integration for service containers (future enhancement)

## Contributing

The Podman support implementation follows these principles:

- **Zero breaking changes** for existing users
- **Runtime-agnostic core** with runtime-specific optimizations
- **Comprehensive error handling** with helpful user guidance
- **Extensive testing** across different environments

See the container package in `pkg/container/` for implementation details.