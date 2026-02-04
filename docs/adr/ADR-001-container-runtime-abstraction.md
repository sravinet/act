# ADR-001: Container Runtime Abstraction

## Status

Implemented ✅ (February 2026)

## Context

Act currently has a tight coupling to Docker through direct Docker API usage throughout the codebase. With the growing adoption of Podman as a daemonless, rootless alternative to Docker, there is increasing demand to support multiple container runtimes while maintaining backward compatibility.

The current architecture has:
- ~184 direct Docker API references across the codebase
- Hard dependencies on Docker client libraries 
- Docker-specific assumptions in networking, volumes, and authentication
- No abstraction between container operations and their implementation

## Decision

We will enhance the existing container abstraction layer to support multiple container runtimes:

1. **Retain existing Container interface**: The current `Container` interface in `pkg/container/container_types.go` already provides a good abstraction
2. **Enhance factory pattern**: Modify `NewContainer()` to detect and select appropriate runtime
3. **Runtime-specific implementations**: Create separate implementations for Docker and Podman while maintaining interface compatibility
4. **Backward compatibility**: Ensure all existing Docker workflows continue unchanged

Key design principles:
- **Zero breaking changes**: Existing users should see no difference
- **Runtime agnostic**: Core act logic should not know about specific runtimes
- **Graceful fallback**: Docker remains the default when both are available
- **Clear error messages**: Users get helpful guidance when runtimes are unavailable

## Consequences

### Positive
- Enables Podman support with minimal risk
- Preserves existing user workflows and configurations
- Sets foundation for future container runtime additions
- Maintains performance characteristics of current implementation
- Leverages existing comprehensive test suite

### Negative
- Increases complexity in container package 
- Requires careful testing across multiple runtime combinations
- Some Docker-specific optimizations may need to be generalized
- Additional maintenance burden for multiple runtime implementations

### Neutral
- Code size will increase modestly (~10-15%)
- Binary size impact minimal due to build tags
- Documentation needs to cover multiple runtime scenarios

## Alternatives Considered

1. **Complete rewrite in Rust**: Rejected due to massive scope and timeline
2. **Fork the project**: Rejected as it fragments the community
3. **Plugin architecture**: Rejected as over-engineered for current needs
4. **Docker compatibility layer only**: Rejected as it doesn't leverage Podman-specific advantages

## Implementation Notes

- Use Go build tags to conditionally compile runtime-specific code
- Maintain separate error handling paths for runtime-specific issues
- Leverage existing socket detection logic in `docker_socket.go`
- Ensure CLI flags support runtime selection and configuration

## Implementation Summary

**Implemented Files:**
- `pkg/container/factory.go` - Enhanced factory pattern with runtime selection
- `pkg/container/runtime_detector.go` - Runtime detection and configuration
- `pkg/container/null_container.go` - Null Object Pattern for graceful error handling
- `cmd/root.go` - CLI configuration integration
- Enhanced `docker_run.go` with runtime-aware methods

**Key Features Delivered:**
- ✅ Zero breaking changes - all existing Docker workflows preserved
- ✅ Runtime-agnostic container factory with intelligent selection
- ✅ Comprehensive error handling with helpful user guidance
- ✅ Extensible architecture for future runtime additions