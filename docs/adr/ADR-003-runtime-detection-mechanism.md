# ADR-003: Runtime Detection Mechanism

## Status

**Implemented ✅ (February 2026)**

**Implementation Notes:**
- Multi-layered detection algorithm implemented with Podman preference → Docker fallback
- CLI integration: `--container-runtime=auto|docker|podman` and `--container-socket=/path`
- Environment variable support with proper precedence (CLI > environment > auto-detection)
- Enhanced macOS support: Automatic Podman machine socket detection via `podman machine inspect`
- Graceful error handling with detailed user guidance when no runtime available
- Socket detection covers all major platforms and installation patterns

## Context

To provide seamless user experience with multiple container runtimes, Act needs an intelligent mechanism to detect, prioritize, and select the appropriate container runtime. Users should not need to configure anything in the common case, but should have control when needed.

Current challenges:
- Docker and Podman may both be installed on the same system
- Socket locations vary across platforms and installations 
- Users may prefer one runtime over another
- Some environments only have one runtime available
- Error messages should guide users when no runtime is available

The existing `socketLocation()` function in `docker_socket.go` already checks multiple socket paths including Podman sockets, but lacks runtime-specific logic.

## Decision

We will implement a **multi-layered runtime detection system** with the following priority order:

### Detection Priority (Highest to Lowest)

1. **Explicit Configuration**: User-specified runtime via CLI or environment
2. **Environment Variables**: Runtime-specific environment variables
3. **Socket Availability + Binary Verification**: Available sockets with working binaries
4. **Auto-detection Fallback**: Intelligent defaults based on platform and availability

### Detection Algorithm

```go
type ContainerRuntime int

const (
    RuntimeUnknown ContainerRuntime = iota
    RuntimeDocker
    RuntimePodman
)

func (rd *RuntimeDetector) DetectAvailableRuntime() ContainerRuntime {
    // 1. Check explicit configuration
    if runtime := rd.checkExplicitConfig(); runtime != RuntimeUnknown {
        return runtime
    }
    
    // 2. Check environment variables
    if runtime := rd.checkEnvironmentHints(); runtime != RuntimeUnknown {
        return runtime
    }
    
    // 3. Auto-detect based on socket + binary availability
    if runtime := rd.autoDetectRuntime(); runtime != RuntimeUnknown {
        return runtime
    }
    
    return RuntimeUnknown
}
```

### Configuration Sources

1. **CLI Flag**: `--container-runtime=podman|docker|auto`
2. **Environment Variable**: `ACT_CONTAINER_RUNTIME=podman|docker|auto`
3. **Socket Override**: `--container-socket=/path/to/socket`
4. **Docker/Podman Environment**: Respect `DOCKER_HOST`, `PODMAN_HOST`

### Auto-Detection Logic

```go
func (rd *RuntimeDetector) autoDetectRuntime() ContainerRuntime {
    runtimes := []RuntimeCandidate{}
    
    // Check each potential runtime
    for _, candidate := range rd.getRuntimeCandidates() {
        if rd.isRuntimeAvailable(candidate) {
            runtimes = append(runtimes, candidate)
        }
    }
    
    // Apply preference logic
    return rd.selectPreferredRuntime(runtimes)
}
```

### Preference Logic

When multiple runtimes are available:
1. **Podman preferred**: Better security (rootless), no daemon required
2. **Docker fallback**: Broader compatibility, mature tooling
3. **User override**: Always respect explicit user choice

### macOS Podman Machine Support

Special handling for Podman on macOS where Podman runs in a VM:
- **Auto-detection**: Uses `podman machine inspect` to get API socket path
- **SSH proxy support**: Detects gvproxy socket that handles SSH tunneling
- **Zero configuration**: Works seamlessly with `podman machine start`

```go
func (rd *RuntimeDetector) getPodmanMachineSocket() (string, bool) {
    cmd := exec.Command("podman", "machine", "inspect", 
        "--format", "{{.ConnectionInfo.PodmanSocket.Path}}")
    output, err := cmd.Output()
    // Returns: /var/folders/.../podman-machine-default-api.sock
}

## Consequences

### Positive
- **Zero-configuration experience**: Works out of the box when possible
- **Predictable behavior**: Clear priority order for runtime selection
- **Flexible override**: Users can control runtime selection when needed
- **Helpful error messages**: Guide users to install or configure runtimes
- **Platform awareness**: Adapts to different OS and installation patterns

### Negative
- **Detection complexity**: More code paths to test and maintain
- **Debug complexity**: Users may be unclear which runtime was selected
- **Performance overhead**: Runtime detection on each invocation
- **Platform differences**: Detection logic may vary across OS

### Risks
- **False positives**: Detecting available runtime that doesn't actually work
- **Priority conflicts**: User expectations may differ from default preferences
- **Environment conflicts**: Multiple runtime installations causing confusion

## Implementation Details

### Socket Detection Enhancement

Extend existing socket detection with runtime identification:

```go
type RuntimeSocket struct {
    Path    string
    Runtime ContainerRuntime
    Score   int // Priority score for selection
}

func detectRuntimeSockets() []RuntimeSocket {
    candidates := []RuntimeSocket{
        {"/var/run/docker.sock", RuntimeDocker, 80},
        {"/run/podman/podman.sock", RuntimePodman, 90},
        {"$XDG_RUNTIME_DIR/podman/podman.sock", RuntimePodman, 95},
        // ... more candidates
    }
    // Return available sockets sorted by score
}
```

### Binary Verification

Verify runtime binaries are functional:

```go
func (rd *RuntimeDetector) verifyRuntime(runtime ContainerRuntime) bool {
    switch runtime {
    case RuntimeDocker:
        return rd.canConnectDocker()
    case RuntimePodman:
        return rd.canConnectPodman()
    }
    return false
}
```

### Error Messaging

Provide helpful guidance when no runtime is available:

```
Error: No container runtime detected

Act requires either Docker or Podman to run GitHub Actions locally.

Install options:
  Docker:  https://docs.docker.com/get-docker/
  Podman:  https://podman.io/getting-started/installation

Current detection status:
  ✗ Docker daemon not running (socket: /var/run/docker.sock)
  ✗ Podman not installed (checked: /run/podman/podman.sock)

Override detection with:
  act --container-runtime=docker
  act --container-socket=/custom/socket
```

## Alternative Approaches Considered

1. **Configuration file based**: Rejected as adds complexity for simple use cases
2. **Runtime auto-install**: Rejected as outside Act's responsibility
3. **Always require explicit configuration**: Rejected as poor user experience
4. **Docker-only with Podman compatibility shim**: Rejected as doesn't leverage Podman advantages

## Future Extensions

This design supports future additions:
- **containerd** support via runtime detection
- **Cloud container services** via API endpoint detection
- **Custom OCI runtimes** via plugin mechanisms
- **Performance caching** of detection results

## Implementation Summary

**Implemented Files:**
- `pkg/container/runtime_detector.go` - Core runtime detection engine
- `pkg/container/runtime_detector_test.go` - Comprehensive test suite
- CLI integration in `cmd/root.go` and `cmd/input.go`
- Runner configuration in `pkg/runner/runner.go`

**Key Features Delivered:**
- ✅ Multi-layered detection with explicit override capability
- ✅ Intelligent socket discovery with priority scoring
- ✅ Environment variable configuration support
- ✅ Helpful error messages with actionable guidance
- ✅ Binary verification to ensure functional runtimes

**Detection Algorithm Implemented:**
1. **Explicit Configuration**: CLI flags and environment variables
2. **Runtime-Specific Hints**: PODMAN_HOST, DOCKER_HOST detection
3. **Socket + Binary Verification**: Available sockets with working binaries
4. **Intelligent Defaults**: Podman preferred, Docker fallback

**User Experience Achievements:**
- Zero-configuration experience for most users
- Clear troubleshooting guidance when no runtime available
- Predictable and documented behavior for all scenarios