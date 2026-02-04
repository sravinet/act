# ADR-004: User Experience Optimizations

## Status

**Implemented ✅ (February 2026)**

**Implementation Notes:**
- Enhanced progress indicators for long-running operations (image pulls)
- Improved logging with clearer status messages and completion feedback
- Performance optimization guidance via CLI flags and documentation
- Quick start guide for common development workflows

## Context

With multi-runtime support implemented, users experienced slow initial runs due to:
- Large container image downloads (1-2GB for common images)
- Multiple parallel jobs pulling the same images simultaneously
- Lack of progress visibility during image operations
- No guidance on performance optimization options

Users needed:
- **Progress visibility**: Clear indication of what's happening during long operations
- **Performance options**: Ways to speed up common development workflows
- **User guidance**: Documentation for optimization and troubleshooting

## Decision

Implement user experience enhancements while maintaining the existing architecture:

### 1. Enhanced Progress Indicators

Add informative logging for image operations:

```go
// Before image pull
logger.Infof("📥 Pulling image '%v' (%s) - this may take a few minutes for large images", imageRef, input.Platform)

// Progress during pull operations
if msg.Status == "Downloading" || msg.Status == "Extracting" || msg.Status == "Pulling fs layer" {
    writeLog(logger, false, "%s %s: %s", msg.Status, msg.ID, msg.Progress)
}

// After completion
logger.Infof("✅ Successfully pulled image '%v'", imageRef)
```

### 2. Performance Optimization Options

Provide CLI flags for common optimization scenarios:

- `--pull=false`: Use cached images only (fastest for development)
- `--jobs N`: Limit parallel jobs to reduce resource contention  
- `-P image=smaller_image`: Use lighter base images
- `--job jobname`: Run specific job for testing

### 3. User Guidance Documentation

Create quick start guide covering:
- Fast development workflows
- Image caching strategies  
- Parallel job optimization
- Troubleshooting common issues

## Implementation

### Enhanced Docker Logger

Modified `pkg/container/docker_logger.go`:
- Show progress bars for download/extract operations
- Highlight important status messages (Pull complete, Already exists)
- Use user-friendly formatting for progress feedback

### Image Pull Enhancements

Modified `pkg/container/docker_pull.go`:
- Add start notification with estimated time guidance
- Show completion confirmation
- Maintain existing caching logic (no performance regression)

### Documentation

Created `QUICK_START.md`:
- Performance tips for common workflows
- CLI flag reference for optimization
- Troubleshooting guide for common issues

## Consequences

### Positive
- **Better user feedback**: Users understand what's happening during long operations
- **Faster development cycles**: Clear guidance on optimization options
- **Reduced support burden**: Self-service troubleshooting documentation
- **Improved adoption**: Better first-run experience

### Negative
- **Additional log output**: More verbose (but informative) logging
- **Documentation maintenance**: Need to keep optimization guide current

### Performance Impact
- **Minimal overhead**: Progress enhancements add negligible performance cost
- **User-controlled optimization**: Users can choose speed vs completeness tradeoffs
- **No regression**: Existing behavior preserved by default

## Examples

### Fast Development Workflow
```bash
# Skip image pulls, use cached images
act --pull=false --container-runtime=podman

# Limit parallel jobs to reduce resource conflicts  
act --jobs 2 --container-runtime=podman

# Use smaller base images
act -P ubuntu-latest=node:16-alpine --container-runtime=podman
```

### Progress Feedback
```
INFO 📥 Pulling image 'catthehacker/ubuntu:act-latest' (linux/amd64) - this may take a few minutes for large images
INFO Downloading sha256:1234: 45.2MB/128.4MB
INFO Extracting sha256:5678: 32.1MB/45.2MB  
INFO Pull complete: sha256:9abc
INFO ✅ Successfully pulled image 'catthehacker/ubuntu:act-latest'
```

## Future Considerations

- **Pull deduplication**: Coordinate multiple jobs pulling same image
- **Image size optimization**: Recommend optimal base images per use case
- **Caching strategies**: More sophisticated image caching policies
- **Progress estimation**: More accurate time estimates for operations