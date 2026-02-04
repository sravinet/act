package container

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/docker/cli/cli/connhelper"
	"github.com/docker/docker/client"
	log "github.com/sirupsen/logrus"
)

var (
	// Global runtime detector instance
	globalDetector *RuntimeDetector
	
	// Allow runtime override for testing
	runtimeOverride ContainerRuntime = RuntimeUnknown
)

// init initializes the global runtime detector
func init() {
	globalDetector = NewRuntimeDetector()
}

// NewContainer creates a reference to a container using the best available runtime
func NewContainer(input *NewContainerInput) ExecutionsEnvironment {
	runtime := getSelectedRuntime()
	
	logger := log.WithFields(log.Fields{
		"component": "container-factory",
		"runtime":   runtime.String(),
	})
	
	switch runtime {
	case RuntimePodman:
		logger.Debug("Creating Podman container")
		return newPodmanContainer(input)
	case RuntimeDocker:
		logger.Debug("Creating Docker container")
		return newDockerContainer(input)
	default:
		logger.Error("No container runtime available")
		// Return a null container that provides helpful error messages
		return newNullContainer(input)
	}
}

// NewContainerWithRuntime creates a container with a specific runtime
func NewContainerWithRuntime(input *NewContainerInput, runtime ContainerRuntime) ExecutionsEnvironment {
	logger := log.WithFields(log.Fields{
		"component": "container-factory",
		"runtime":   runtime.String(),
		"forced":    true,
	})
	
	// Verify the requested runtime is available
	if !globalDetector.verifyRuntime(runtime) {
		logger.Errorf("Requested runtime %s is not available", runtime.String())
		return newNullContainer(input)
	}
	
	switch runtime {
	case RuntimePodman:
		logger.Debug("Creating Podman container (forced)")
		return newPodmanContainer(input)
	case RuntimeDocker:
		logger.Debug("Creating Docker container (forced)")
		return newDockerContainer(input)
	default:
		logger.Error("Invalid runtime specified")
		return newNullContainer(input)
	}
}

// getSelectedRuntime determines which runtime to use
func getSelectedRuntime() ContainerRuntime {
	// Check for testing override
	if runtimeOverride != RuntimeUnknown {
		return runtimeOverride
	}
	
	// Apply any configuration from environment or CLI
	configureDetectorFromEnvironment()
	
	// Detect available runtime
	return globalDetector.DetectAvailableRuntime()
}

// configureDetectorFromEnvironment applies environment-based configuration
func configureDetectorFromEnvironment() {
	// Check for runtime preference
	if runtime := os.Getenv("ACT_CONTAINER_RUNTIME"); runtime != "" {
		switch runtime {
		case "docker":
			globalDetector.SetPreferredRuntime(RuntimeDocker)
		case "podman":
			globalDetector.SetPreferredRuntime(RuntimePodman)
		}
	}
	
	// Check for custom socket
	if socket := os.Getenv("ACT_CONTAINER_SOCKET"); socket != "" {
		globalDetector.SetCustomSocket(socket)
	}
}

// SetRuntimePreference sets the global runtime preference (for CLI configuration)
func SetRuntimePreference(runtime ContainerRuntime) {
	globalDetector.SetPreferredRuntime(runtime)
}

// SetCustomSocket sets a custom socket path (for CLI configuration)
func SetCustomSocket(socket string) {
	globalDetector.SetCustomSocket(socket)
}

// GetCurrentRuntime returns the currently selected runtime without creating a container
func GetCurrentRuntime() ContainerRuntime {
	return getSelectedRuntime()
}

// GetAvailableRuntimes returns all detected available runtimes
func GetAvailableRuntimes() []ContainerRuntime {
	var available []ContainerRuntime
	
	if globalDetector.verifyRuntime(RuntimeDocker) {
		available = append(available, RuntimeDocker)
	}
	
	if globalDetector.verifyRuntime(RuntimePodman) {
		available = append(available, RuntimePodman)
	}
	
	return available
}

// GetRuntimeDetectionError returns a helpful error message when no runtime is available
func GetRuntimeDetectionError() string {
	return globalDetector.GetHelpfulErrorMessage()
}

// SetRuntimeOverride sets a runtime override for testing
func SetRuntimeOverride(runtime ContainerRuntime) {
	runtimeOverride = runtime
}

// ClearRuntimeOverride clears the runtime override
func ClearRuntimeOverride() {
	runtimeOverride = RuntimeUnknown
}

// GetContainerClient returns a runtime-aware container client
// This replaces GetDockerClient() with multi-runtime support
func GetContainerClient(ctx context.Context) (client.APIClient, error) {
	runtime := getSelectedRuntime()
	
	logger := log.WithFields(log.Fields{
		"component": "container-client",
		"runtime":   runtime.String(),
	})
	
	switch runtime {
	case RuntimePodman:
		return createPodmanClient(ctx, logger)
	case RuntimeDocker:
		return createDockerClient(ctx, logger)
	default:
		return nil, fmt.Errorf("no container runtime available\n\n%s", GetRuntimeDetectionError())
	}
}

// createDockerClient creates a Docker client (replaces the old GetDockerClient logic)
func createDockerClient(ctx context.Context, logger *log.Entry) (client.APIClient, error) {
	logger.Debug("Creating Docker client")
	
	dockerHost := os.Getenv("DOCKER_HOST")
	
	var cli client.APIClient
	var err error
	
	if strings.HasPrefix(dockerHost, "ssh://") {
		var helper *connhelper.ConnectionHelper
		helper, err = connhelper.GetConnectionHelper(dockerHost)
		if err != nil {
			return nil, fmt.Errorf("failed to create SSH connection helper: %w", err)
		}
		cli, err = client.NewClientWithOpts(
			client.WithHost(helper.Host),
			client.WithDialContext(helper.Dialer),
		)
	} else {
		cli, err = client.NewClientWithOpts(client.FromEnv)
	}
	
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Docker daemon: %w", err)
	}
	
	cli.NegotiateAPIVersion(ctx)
	logger.Debug("Successfully connected to Docker daemon")
	return cli, nil
}

// createPodmanClient creates a Podman client using Docker-compatible API
func createPodmanClient(ctx context.Context, logger *log.Entry) (client.APIClient, error) {
	logger.Debug("Creating Podman client")
	
	// Get Podman socket from our runtime detector
	socket, found := globalDetector.GetSocketForRuntime(RuntimePodman)
	if !found {
		return nil, fmt.Errorf("podman socket not found or not accessible")
	}
	
	logger.Debugf("Connecting to Podman at %s", socket)
	
	var cli client.APIClient
	var err error
	
	// Check if it's an SSH connection
	if strings.HasPrefix(socket, "ssh://") {
		var helper *connhelper.ConnectionHelper
		helper, err = connhelper.GetConnectionHelper(socket)
		if err != nil {
			return nil, fmt.Errorf("failed to create SSH connection helper for Podman: %w", err)
		}
		cli, err = client.NewClientWithOpts(
			client.WithHost(helper.Host),
			client.WithDialContext(helper.Dialer),
		)
	} else {
		// Direct socket connection
		cli, err = client.NewClientWithOpts(
			client.WithHost(socket),
			client.WithAPIVersionNegotiation(),
		)
	}
	
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Podman daemon: %w", err)
	}
	
	// Verify connection works
	if _, err = cli.Ping(ctx); err != nil {
		cli.Close()
		return nil, fmt.Errorf("failed to ping Podman daemon: %w", err)
	}
	
	logger.Debug("Successfully connected to Podman daemon")
	return cli, nil
}