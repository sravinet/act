package container

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/docker/docker/client"
	log "github.com/sirupsen/logrus"
)

// ContainerRuntime represents the detected container runtime
type ContainerRuntime int

const (
	RuntimeUnknown ContainerRuntime = iota
	RuntimeDocker
	RuntimePodman
)

// String returns the string representation of the runtime
func (r ContainerRuntime) String() string {
	switch r {
	case RuntimeDocker:
		return "docker"
	case RuntimePodman:
		return "podman"
	default:
		return "unknown"
	}
}

// RuntimeSocket represents a detected container runtime socket
type RuntimeSocket struct {
	Path    string
	Runtime ContainerRuntime
	Score   int // Priority score for selection (higher is better)
}

// RuntimeDetector handles detection and selection of container runtimes
type RuntimeDetector struct {
	preferredRuntime ContainerRuntime
	customSocket     string
	logger           *log.Entry
}

// NewRuntimeDetector creates a new runtime detector
func NewRuntimeDetector() *RuntimeDetector {
	return &RuntimeDetector{
		preferredRuntime: RuntimeUnknown,
		logger:           log.WithField("component", "runtime-detector"),
	}
}

// SetPreferredRuntime sets the preferred runtime (overrides auto-detection)
func (rd *RuntimeDetector) SetPreferredRuntime(runtime ContainerRuntime) {
	rd.preferredRuntime = runtime
	rd.logger.Debugf("Preferred runtime set to: %s", runtime.String())
}

// SetCustomSocket sets a custom socket path (overrides detection)
func (rd *RuntimeDetector) SetCustomSocket(socket string) {
	rd.customSocket = socket
	rd.logger.Debugf("Custom socket set to: %s", socket)
}

// DetectAvailableRuntime detects and returns the best available container runtime
func (rd *RuntimeDetector) DetectAvailableRuntime() ContainerRuntime {
	rd.logger.Debug("Starting container runtime detection")

	// 1. Check explicit configuration
	if rd.preferredRuntime != RuntimeUnknown {
		if rd.verifyRuntime(rd.preferredRuntime) {
			rd.logger.Infof("Using preferred runtime: %s", rd.preferredRuntime.String())
			return rd.preferredRuntime
		}
		rd.logger.Warnf("Preferred runtime %s is not available, falling back to auto-detection", rd.preferredRuntime.String())
	}

	// 2. Check environment variables
	if runtime := rd.checkEnvironmentHints(); runtime != RuntimeUnknown {
		if rd.verifyRuntime(runtime) {
			rd.logger.Infof("Using runtime from environment: %s", runtime.String())
			return runtime
		}
		rd.logger.Warnf("Environment-specified runtime %s is not available", runtime.String())
	}

	// 3. Auto-detect based on socket + binary availability
	if runtime := rd.autoDetectRuntime(); runtime != RuntimeUnknown {
		rd.logger.Infof("Auto-detected runtime: %s", runtime.String())
		return runtime
	}

	rd.logger.Error("No container runtime detected")
	return RuntimeUnknown
}

// checkEnvironmentHints checks environment variables for runtime hints
func (rd *RuntimeDetector) checkEnvironmentHints() ContainerRuntime {
	// Check ACT-specific environment variable
	if actRuntime := os.Getenv("ACT_CONTAINER_RUNTIME"); actRuntime != "" {
		switch strings.ToLower(actRuntime) {
		case "docker":
			return RuntimeDocker
		case "podman":
			return RuntimePodman
		}
	}

	// Check for Podman-specific environment variables
	if os.Getenv("PODMAN_HOST") != "" {
		return RuntimePodman
	}

	// Check for Docker-specific environment variables
	if os.Getenv("DOCKER_HOST") != "" {
		return RuntimeDocker
	}

	return RuntimeUnknown
}

// autoDetectRuntime performs automatic runtime detection
func (rd *RuntimeDetector) autoDetectRuntime() ContainerRuntime {
	sockets := rd.detectRuntimeSockets()
	if len(sockets) == 0 {
		return RuntimeUnknown
	}

	// Try each socket in order of priority
	for _, socket := range sockets {
		if rd.verifySocketConnection(socket) {
			return socket.Runtime
		}
	}

	return RuntimeUnknown
}

// detectRuntimeSockets finds available container runtime sockets
func (rd *RuntimeDetector) detectRuntimeSockets() []RuntimeSocket {
	var available []RuntimeSocket

	// If custom socket is specified, only try that
	if rd.customSocket != "" {
		runtime := rd.guessRuntimeFromSocket(rd.customSocket)
		return []RuntimeSocket{{
			Path:    rd.customSocket,
			Runtime: runtime,
			Score:   100,
		}}
	}

	// Check common socket locations
	candidates := []RuntimeSocket{
		// Podman sockets (preferred for security/performance)
		{"$XDG_RUNTIME_DIR/podman/podman.sock", RuntimePodman, 95},
		{"/run/podman/podman.sock", RuntimePodman, 90},
		{"$HOME/.local/share/containers/podman/machine/podman.sock", RuntimePodman, 85},

		// Docker sockets
		{"/var/run/docker.sock", RuntimeDocker, 80},
		{"$HOME/.colima/docker.sock", RuntimeDocker, 75},
		{"$XDG_RUNTIME_DIR/docker.sock", RuntimeDocker, 70},
		{"$HOME/.docker/run/docker.sock", RuntimeDocker, 65},

		// Windows named pipes
		{`\\.\pipe\docker_engine`, RuntimeDocker, 60},
		{`\\.\pipe\podman-machine-default`, RuntimePodman, 85},
	}

	for _, candidate := range candidates {
		expanded := os.ExpandEnv(candidate.Path)
		if rd.socketExists(expanded) {
			available = append(available, RuntimeSocket{
				Path:    expanded,
				Runtime: candidate.Runtime,
				Score:   candidate.Score,
			})
		}
	}

	// Sort by score (highest first)
	for i := 0; i < len(available)-1; i++ {
		for j := i + 1; j < len(available); j++ {
			if available[i].Score < available[j].Score {
				available[i], available[j] = available[j], available[i]
			}
		}
	}

	return available
}

// socketExists checks if a socket file exists
func (rd *RuntimeDetector) socketExists(path string) bool {
	if stat, err := os.Lstat(path); err == nil {
		// Check if it's a socket or named pipe
		mode := stat.Mode()
		return mode&os.ModeSocket != 0 || mode&os.ModeNamedPipe != 0
	}
	return false
}

// guessRuntimeFromSocket attempts to guess runtime from socket path
func (rd *RuntimeDetector) guessRuntimeFromSocket(socket string) ContainerRuntime {
	lower := strings.ToLower(socket)
	if strings.Contains(lower, "podman") {
		return RuntimePodman
	}
	if strings.Contains(lower, "docker") {
		return RuntimeDocker
	}
	// Default to Docker for unknown sockets (Docker-compatible API assumption)
	return RuntimeDocker
}

// verifyRuntime verifies that a runtime is available and functional
func (rd *RuntimeDetector) verifyRuntime(runtime ContainerRuntime) bool {
	switch runtime {
	case RuntimeDocker:
		return rd.verifyDocker()
	case RuntimePodman:
		return rd.verifyPodman()
	default:
		return false
	}
}

// verifySocketConnection verifies a specific socket can be connected to
func (rd *RuntimeDetector) verifySocketConnection(socket RuntimeSocket) bool {
	var host string
	if strings.HasPrefix(socket.Path, `\\.\`) {
		host = "npipe://" + filepath.ToSlash(socket.Path)
	} else {
		host = "unix://" + socket.Path
	}

	// Create a temporary client to test the connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cli, err := client.NewClientWithOpts(
		client.WithHost(host),
		client.WithAPIVersionNegotiation(),
	)
	if err != nil {
		rd.logger.Debugf("Failed to create client for %s: %v", socket.Path, err)
		return false
	}
	defer cli.Close()

	// Try to ping the daemon
	_, err = cli.Ping(ctx)
	if err != nil {
		rd.logger.Debugf("Failed to ping daemon at %s: %v", socket.Path, err)
		return false
	}

	rd.logger.Debugf("Successfully verified %s runtime at %s", socket.Runtime.String(), socket.Path)
	return true
}

// verifyDocker checks if Docker is available and functional
func (rd *RuntimeDetector) verifyDocker() bool {
	// Check if docker binary is available
	if _, err := exec.LookPath("docker"); err != nil {
		rd.logger.Debug("Docker binary not found in PATH")
		return false
	}

	// Try to connect via existing socket detection
	if socket, found := socketLocation(); found {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		cli, err := client.NewClientWithOpts(
			client.WithHost(socket),
			client.WithAPIVersionNegotiation(),
		)
		if err != nil {
			rd.logger.Debugf("Failed to create Docker client: %v", err)
			return false
		}
		defer cli.Close()

		_, err = cli.Ping(ctx)
		return err == nil
	}

	return false
}

// verifyPodman checks if Podman is available and functional
func (rd *RuntimeDetector) verifyPodman() bool {
	// Check if podman binary is available
	if _, err := exec.LookPath("podman"); err != nil {
		rd.logger.Debug("Podman binary not found in PATH")
		return false
	}

	// Try podman info command as a quick check
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "podman", "info", "--format", "json")
	if err := cmd.Run(); err != nil {
		rd.logger.Debugf("Podman info command failed: %v", err)
		return false
	}

	return true
}

// getPodmanMachineSocket gets the Podman machine API socket path on macOS
func (rd *RuntimeDetector) getPodmanMachineSocket() (string, bool) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	
	cmd := exec.CommandContext(ctx, "podman", "machine", "inspect", "--format", "{{.ConnectionInfo.PodmanSocket.Path}}")
	output, err := cmd.Output()
	if err != nil {
		rd.logger.Debugf("Failed to get Podman machine socket: %v", err)
		return "", false
	}
	
	socketPath := strings.TrimSpace(string(output))
	if socketPath == "" || socketPath == "<no value>" {
		rd.logger.Debug("No Podman machine socket path found")
		return "", false
	}
	
	// Verify socket exists and is accessible
	if _, err := os.Stat(socketPath); err != nil {
		rd.logger.Debugf("Podman machine socket not accessible: %v", err)
		return "", false
	}
	
	rd.logger.Debugf("Found Podman machine socket: %s", socketPath)
	return socketPath, true
}

// GetSocketForRuntime returns the socket path for a specific runtime
func (rd *RuntimeDetector) GetSocketForRuntime(runtime ContainerRuntime) (string, bool) {
	if rd.customSocket != "" {
		return rd.customSocket, true
	}

	// Special handling for Podman on macOS - check for machine socket first
	if runtime == RuntimePodman {
		if socketPath, found := rd.getPodmanMachineSocket(); found {
			return "unix://" + socketPath, true
		}
	}

	sockets := rd.detectRuntimeSockets()
	for _, socket := range sockets {
		if socket.Runtime == runtime && rd.verifySocketConnection(socket) {
			var socketURI string
			if strings.HasPrefix(socket.Path, `\\.\`) {
				socketURI = "npipe://" + filepath.ToSlash(socket.Path)
			} else {
				socketURI = "unix://" + socket.Path
			}
			return socketURI, true
		}
	}

	return "", false
}

// GetHelpfulErrorMessage returns a user-friendly error message when no runtime is available
func (rd *RuntimeDetector) GetHelpfulErrorMessage() string {
	var message strings.Builder
	
	message.WriteString("No container runtime detected\n\n")
	message.WriteString("Act requires either Docker or Podman to run GitHub Actions locally.\n\n")
	message.WriteString("Install options:\n")
	message.WriteString("  Docker:  https://docs.docker.com/get-docker/\n")
	message.WriteString("  Podman:  https://podman.io/getting-started/installation\n\n")
	
	message.WriteString("Current detection status:\n")
	
	// Check Docker
	dockerAvailable := rd.verifyDocker()
	dockerSocket, dockerSocketFound := socketLocation()
	dockerStatus := "✗"
	if dockerAvailable {
		dockerStatus = "✓"
	}
	
	if dockerSocketFound {
		message.WriteString(fmt.Sprintf("  %s Docker (socket: %s)\n", dockerStatus, dockerSocket))
	} else {
		message.WriteString("  ✗ Docker daemon not running (no socket found)\n")
	}
	
	// Check Podman
	podmanAvailable := rd.verifyPodman()
	podmanStatus := "✗"
	if podmanAvailable {
		podmanStatus = "✓"
	}
	message.WriteString(fmt.Sprintf("  %s Podman (binary check)\n", podmanStatus))
	
	message.WriteString("\nOverride detection with:\n")
	message.WriteString("  act --container-runtime=docker\n")
	message.WriteString("  act --container-runtime=podman\n")
	message.WriteString("  act --container-socket=/custom/socket\n")
	
	return message.String()
}