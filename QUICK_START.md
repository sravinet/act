# 🚀 Act Quick Start & Performance Tips

## Fast Development Workflow

### 1. **Skip Image Pulls (Use Cached)**
```bash
# Only use locally cached images
~/bin/act --pull=false

# Use smaller, faster images
~/bin/act -P ubuntu-latest=node:16-alpine
```

### 2. **Limit Parallel Jobs**
```bash
# Reduce resource contention
~/bin/act --jobs 2

# Run single job for testing
~/bin/act --job lint
```

### 3. **Pre-pull Images**
```bash
# Pull common images ahead of time
podman pull catthehacker/ubuntu:act-latest
podman pull node:16-alpine
podman pull ubuntu:latest
```

### 4. **Container Runtime Options**
```bash
# Use Podman (faster, more secure)
~/bin/act --container-runtime=podman

# Use Docker (if Podman issues)
~/bin/act --container-runtime=docker

# Auto-detect best runtime
~/bin/act --container-runtime=auto
```

### 5. **Debugging & Progress**
```bash
# Verbose output (shows progress)
~/bin/act --verbose

# Dry run (no actual execution)
~/bin/act --dryrun

# List available jobs
~/bin/act --list
```

## Performance Tips

- **First run**: May take 2-5 minutes (image download)
- **Subsequent runs**: Much faster (cached images)
- **Large images**: Use `--jobs 1` to avoid parallel download conflicts
- **CI/CD**: Consider using smaller base images

## Troubleshooting

```bash
# Check what's available
~/bin/act --list

# Test single job
~/bin/act --job test --dryrun

# Force clean pull
~/bin/act --pull=true --job test
```