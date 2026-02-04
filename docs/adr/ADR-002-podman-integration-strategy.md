# ADR-002: Podman Integration Strategy

## Status

**Implemented ✅ (February 2026)**

**Implementation Notes:**
- Docker-compatible API strategy successfully implemented via Docker client libraries
- Runtime-specific optimizations added for Podman (rootless detection, startup improvements)
- Factory pattern enables seamless switching between Docker and Podman implementations
- All container operations (build, pull, run, network, volume) support both runtimes

## Context

With the decision to support multiple container runtimes, we need to determine the best approach for integrating Podman support. Podman offers a Docker-compatible REST API, but has different characteristics:

- **Daemonless architecture**: No background daemon required
- **Rootless by default**: Improved security posture
- **systemd integration**: Native systemd support for container management
- **Different networking**: Uses slirp4netns and other rootless networking approaches
- **Podman REST API**: Docker-compatible API with some extensions

Key considerations:
- Podman provides Docker-compatible API endpoints
- Different socket locations and discovery mechanisms
- Rootless containers have different volume mounting behaviors
- Network management differs significantly from Docker

## Decision

We will implement Podman support using the **Docker-compatible client approach**:

1. **Leverage Docker client libraries**: Use existing `github.com/docker/docker/client` with Podman sockets
2. **Runtime-specific adaptations**: Handle Podman-specific networking, volumes, and authentication 
3. **Socket-based detection**: Extend existing socket detection to identify Podman sockets
4. **Conditional behavior**: Adapt specific operations based on detected runtime

### Implementation Strategy

```go
// Use Docker client with Podman socket
func newPodmanContainer(input *NewContainerInput) ExecutionsEnvironment {
    return &containerReference{
        input: input,
        runtime: RuntimePodman,
        // Docker client connected to Podman socket
    }
}
```

### Podman-Specific Adaptations

1. **Networking**: Handle rootless networking constraints
2. **Volumes**: Adapt for rootless volume mounting behaviors  
3. **Authentication**: Support Podman-specific registry authentication
4. **Health checks**: Adapt health checking for Podman containers
5. **Process management**: Handle daemonless container lifecycle

## Consequences

### Positive
- **Rapid implementation**: Leverage existing Docker client infrastructure
- **Proven compatibility**: Podman's Docker-compatible API is well-tested
- **Code reuse**: Minimal duplication of container management logic
- **Familiar patterns**: Developers can apply Docker knowledge to Podman support

### Negative
- **API limitations**: Some Podman-specific features won't be accessible via Docker API
- **Error translation**: Podman errors need translation to user-friendly messages
- **Performance overhead**: Small overhead from Docker API compatibility layer

### Risks
- **API drift**: Podman's Docker compatibility may change over time
- **Feature gaps**: Some advanced Podman features require native API usage
- **Debugging complexity**: Error messages may be less clear than native Podman errors

## Alternatives Considered

1. **Native Podman client**: 
   - Pros: Full access to Podman features, better error messages
   - Cons: Significant development effort, code duplication
   - Decision: Rejected due to implementation complexity

2. **Podman REST API directly**:
   - Pros: Modern REST interface, full feature access
   - Cons: Additional HTTP client complexity, different patterns
   - Decision: Rejected for initial implementation

3. **CLI wrapper approach**:
   - Pros: Simple to implement, direct access to all features
   - Cons: Performance overhead, complex error handling, process management
   - Decision: Rejected due to reliability concerns

## Implementation Plan

### Phase 1: Foundation
- Extend runtime detection to identify Podman installations
- Modify factory to create Podman-aware containers
- Basic container lifecycle operations (create, start, stop, remove)

### Phase 2: Core Functionality  
- Image operations (pull, build) with Podman
- Volume mounting with rootless considerations
- Environment variable and secret management

### Phase 3: Advanced Features
- Podman-specific networking configuration
- Health check adaptation for Podman containers
- Performance optimizations for rootless scenarios

### Future Considerations
If Docker compatibility proves insufficient, we can:
- Implement native Podman client as alternative
- Provide hybrid approach with runtime-specific optimizations
- Extend to other OCI-compatible runtimes

## Implementation Summary

**Implemented Files:**
- `pkg/container/podman_run.go` - Podman-specific container implementation
- Enhanced socket detection in `runtime_detector.go` for Podman socket discovery
- Runtime-specific error handling and user guidance

**Key Features Delivered:**
- ✅ Docker-compatible API usage for seamless integration
- ✅ Podman-specific error detection and helpful hints
- ✅ Rootless container detection and optimization hooks
- ✅ Graceful fallback between runtimes
- ✅ Production-ready implementation with comprehensive testing

**Podman Integration Success Metrics:**
- Zero learning curve for existing Act users
- Automatic runtime detection with Podman preferred for security
- Full feature parity with Docker implementation
- Enhanced security through rootless container support