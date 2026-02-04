# ADR-005: Apple Container Runtime Evaluation

## Status

**Evaluated ✅ (February 2026) - Deferred for Future Implementation**

**Decision Summary:**
- Apple Container offers significant technical advantages but is too early for production adoption
- Current Podman/Docker multi-runtime architecture provides excellent coverage for immediate needs
- Architecture prepared for future Apple Container integration when mature
- Regular re-evaluation planned as Apple Container ecosystem develops

## Context

Apple has released an official container runtime called "Apple Container" (also known as `container`) that provides:

- **Native Apple Silicon optimization**: Written in Swift, optimized for Apple hardware
- **Enhanced security model**: Each container runs in its own lightweight VM using the Virtualization framework
- **Deep macOS integration**: Native integration with vmnet, XPC, Launchd, Keychain services
- **OCI compatibility**: Consumes and produces standard OCI container images
- **Performance benefits**: Lightweight VMs with fast boot times and lower memory usage

### Current Act Runtime Support

Act currently supports:
1. **Docker**: Mature, widely adopted, Docker-compatible API
2. **Podman**: Rootless, daemonless, Docker-compatible API with enhanced security
3. **Multi-runtime detection**: Automatic detection with user override capabilities

### The Question

Should Act add Apple Container as a third runtime option for enhanced macOS native performance and security?

## Technical Analysis

### Apple Container Strengths

**Security & Isolation:**
- **VM-level isolation**: Each container runs in its own lightweight VM
- **Reduced attack surface**: Minimal core utilities and dynamic libraries
- **Privacy protection**: Mount only necessary data into each VM
- **Native sandboxing**: Leverages macOS security frameworks

**Performance:**
- **Apple Silicon optimization**: Native Swift implementation for Apple hardware
- **Efficient resource usage**: Lower memory overhead than traditional VMs
- **Fast boot times**: Comparable to containers in shared VM
- **macOS framework integration**: Direct use of Virtualization and vmnet frameworks

**Developer Experience:**
- **OCI compatibility**: Works with existing container images and workflows
- **Native tooling**: Integrated with macOS development environment
- **Keychain integration**: Seamless registry authentication

### Implementation Challenges

**API Compatibility:**
- **No Docker API**: Uses custom Swift-based XPC API, not REST-compatible
- **Custom integration required**: Would need significant implementation work
- **Different command structure**: Commands don't map 1:1 with Docker/Podman

**Maturity & Adoption:**
- **Very new**: First released in 2025, still in active development
- **API stability**: Interfaces may change as project matures
- **Limited adoption**: Not yet widely used in CI/CD ecosystem
- **macOS version requirement**: Requires macOS 26+ (not widely available)

**Ecosystem Integration:**
- **Limited tooling**: Fewer third-party tools support Apple Container
- **CI/CD compatibility**: Most systems expect Docker-compatible APIs
- **Image registry support**: While OCI-compatible, tooling may not be optimized

## Decision

**Defer Apple Container integration** for the following reasons:

### Current State Assessment

1. **Timing**: macOS 26 requirement limits immediate adoption
2. **API Stability**: Too early in development cycle for production dependency
3. **Ecosystem Maturity**: Limited tooling and CI/CD integration
4. **Current Solution Quality**: Podman/Docker multi-runtime already provides excellent coverage

### Strategic Approach

1. **Monitor Development**: Regular evaluation of Apple Container maturity
2. **Architecture Readiness**: Ensure multi-runtime architecture can accommodate future addition
3. **Community Feedback**: Gauge user interest and demand for Apple Container support
4. **Apple Adoption**: Wait for broader Apple ecosystem adoption signals

### Future Integration Path

When appropriate, Apple Container can be added as `RuntimeApple`:

```go
// Future implementation outline
case RuntimeApple:
    return newAppleContainer(input)

func newAppleContainer(input *NewContainerInput) ExecutionsEnvironment {
    // Custom implementation using Apple Container XPC APIs
    // Would require significant adapter layer for Docker API compatibility
}
```

## Implementation Considerations

### Architecture Compatibility

**Existing Factory Pattern Ready:**
```go
// Current runtime detection easily extended
func (rd *RuntimeDetector) autoDetectRuntime() ContainerRuntime {
    // Priority order when mature:
    // 1. Check for Apple Container (macOS native)
    // 2. Check for Podman (security-focused)
    // 3. Check for Docker (broad compatibility)
}
```

**Configuration Integration:**
```bash
# Future CLI integration
act --container-runtime=apple
ACT_CONTAINER_RUNTIME=apple act
```

### Required Implementation Work

**Major Components:**
1. **Apple Container Client**: XPC-based client for container operations
2. **API Adapter**: Bridge Apple Container APIs to Act's container interface
3. **Detection Logic**: Runtime detection and verification for Apple Container
4. **Command Mapping**: Map Docker/Podman commands to Apple Container equivalents

**Estimated Effort:**
- **High**: Requires custom XPC integration and API adaptation
- **Timeline**: 4-6 weeks of development when ready
- **Risk**: API changes during Apple Container development

## Consequences

### Positive

**Future Opportunities:**
- **Native Performance**: Best possible container performance on Apple Silicon
- **Enhanced Security**: VM-level isolation superior to namespace-based containers
- **Apple Ecosystem**: First-class macOS integration and developer experience
- **Innovation Leadership**: Early adoption of cutting-edge container technology

**Strategic Benefits:**
- **Architecture Flexibility**: Multi-runtime design accommodates future additions
- **Competitive Advantage**: Unique native macOS container support when mature
- **Community Value**: Provide best-in-class tooling for Apple developers

### Negative

**Current Limitations:**
- **Additional Complexity**: Another runtime to test, document, and support
- **Limited User Base**: Initially small audience due to macOS 26 requirement
- **Maintenance Burden**: Custom integration requires ongoing maintenance
- **API Evolution Risk**: Breaking changes during Apple Container development

### Mitigation Strategies

**Monitoring Plan:**
- **Quarterly Review**: Regular assessment of Apple Container maturity
- **Community Engagement**: Track user requests and interest levels
- **Apple Signals**: Monitor Apple's investment and ecosystem development
- **Technical Readiness**: Prepare implementation roadmap for rapid deployment

## Review Timeline

### Immediate Actions (2026)
- ✅ Document opportunity in ADR-005
- ✅ Ensure architecture supports future extension
- 📋 Monitor Apple Container development and adoption

### Short-term Review (Q3 2026)
- 🔍 Assess macOS 26 adoption rates
- 📊 Evaluate Apple Container API stability
- 👥 Gauge community interest and requests
- 🎯 Review competitive landscape

### Medium-term Evaluation (Q1 2027)
- 📈 Analyze Apple Container ecosystem maturity
- 🛠️ Assess tooling and CI/CD integration progress
- 💼 Consider enterprise adoption patterns
- 🚀 Make implementation decision if conditions met

## Examples

### Potential Future Usage

```bash
# Automatic detection (Apple Container preferred on macOS 26+)
act --container-runtime=auto

# Explicit Apple Container usage
act --container-runtime=apple

# Mixed environment support
act --container-runtime=apple --job test-macos
act --container-runtime=podman --job test-linux
```

### Architecture Integration

```go
// Future runtime detection enhancement
func (rd *RuntimeDetector) detectAppleContainer() bool {
    // Check for macOS version
    if !rd.isMacOS26OrLater() {
        return false
    }
    
    // Check for container binary and XPC service
    if _, err := exec.LookPath("container"); err != nil {
        return false
    }
    
    // Verify XPC service availability
    return rd.verifyAppleContainerService()
}
```

## References

- [Apple Container GitHub Repository](https://github.com/apple/container)
- [Apple Containerization Package](https://github.com/apple/containerization)
- [macOS Virtualization Framework](https://developer.apple.com/documentation/virtualization)
- [OCI Container Specifications](https://github.com/opencontainers/image-spec)

## Decision Authors

- System Architecture Team
- Container Runtime Working Group

## Next Review Date

**Q3 2026** - Reassess when macOS 26 reaches broader adoption and Apple Container APIs stabilize