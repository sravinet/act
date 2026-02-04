package container

import (
	"os"

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