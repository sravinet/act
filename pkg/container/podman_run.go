//go:build !(WITHOUT_DOCKER || !(linux || darwin || windows || netbsd))

package container

import (
	"context"
	"fmt"
	"strings"

	"github.com/docker/docker/client"

	"github.com/nektos/act/pkg/common"
)

// newPodmanContainer creates a reference to a Podman container using Docker-compatible API
func newPodmanContainer(input *NewContainerInput) ExecutionsEnvironment {
	cr := new(containerReference)
	cr.input = input
	cr.runtime = RuntimePodman
	return cr
}

// connectPodman creates a connection to the Podman daemon using Docker-compatible API
func (cr *containerReference) connectPodman() common.Executor {
	return func(ctx context.Context) error {
		if cr.cli != nil {
			return nil
		}

		logger := common.Logger(ctx)
		
		// Get Podman socket
		socket, found := globalDetector.GetSocketForRuntime(RuntimePodman)
		if !found {
			return fmt.Errorf("podman socket not found or not accessible")
		}

		logger.Debugf("Connecting to Podman at %s", socket)

		// Create Docker-compatible client connected to Podman socket
		cli, err := client.NewClientWithOpts(
			client.WithHost(socket),
			client.WithAPIVersionNegotiation(),
		)
		if err != nil {
			return fmt.Errorf("failed to create Podman client: %w", err)
		}

		// Verify connection
		if _, err = cli.Ping(ctx); err != nil {
			cli.Close()
			return fmt.Errorf("failed to ping Podman daemon: %w", err)
		}

		cr.cli = cli
		return nil
	}
}

// createPodman creates a Podman container with Podman-specific optimizations
func (cr *containerReference) createPodman(capAdd []string, capDrop []string) common.Executor {
	return func(ctx context.Context) error {
		logger := common.Logger(ctx)
		
		// Use the existing create logic but with Podman-specific adaptations
		if err := cr.createGeneric(capAdd, capDrop)(ctx); err != nil {
			// Check for Podman-specific errors and provide better messages
			if isPodmanSpecificError(err) {
				return fmt.Errorf("podman container creation failed: %w\nHint: %s", err, getPodmanErrorHint(err))
			}
			return err
		}

		logger.Debugf("Successfully created Podman container: %s", cr.id)
		return nil
	}
}

// startPodman starts a Podman container with Podman-specific handling
func (cr *containerReference) startPodman() common.Executor {
	return func(ctx context.Context) error {
		logger := common.Logger(ctx)
		
		// Check for rootless Podman considerations
		if isRootlessPodman(ctx, cr.cli) {
			logger.Debug("Detected rootless Podman, applying rootless optimizations")
			if err := cr.applyRootlessOptimizations(ctx); err != nil {
				logger.Warnf("Failed to apply rootless optimizations: %v", err)
			}
		}

		// Use standard start logic
		return cr.startGeneric()(ctx)
	}
}

// applyRootlessOptimizations applies optimizations specific to rootless Podman
func (cr *containerReference) applyRootlessOptimizations(ctx context.Context) error {
	// For rootless containers, we may need to adjust:
	// 1. User namespace mappings
	// 2. Volume mount options
	// 3. Network configurations
	
	logger := common.Logger(ctx)
	logger.Debug("Applying rootless Podman optimizations")
	
	// This is a placeholder for future rootless-specific optimizations
	// For now, the Docker-compatible API handles most cases correctly
	
	return nil
}

// isPodmanSpecificError checks if an error is Podman-specific
func isPodmanSpecificError(err error) bool {
	if err == nil {
		return false
	}
	
	errStr := strings.ToLower(err.Error())
	podmanIndicators := []string{
		"slirp4netns",
		"rootless",
		"user namespace",
		"podman",
	}
	
	for _, indicator := range podmanIndicators {
		if strings.Contains(errStr, indicator) {
			return true
		}
	}
	
	return false
}

// getPodmanErrorHint provides helpful hints for Podman-specific errors
func getPodmanErrorHint(err error) string {
	if err == nil {
		return ""
	}
	
	errStr := strings.ToLower(err.Error())
	
	if strings.Contains(errStr, "rootless") {
		return "This may be related to rootless Podman. Try running with --privileged or check user namespace configuration."
	}
	
	if strings.Contains(errStr, "slirp4netns") {
		return "This may be related to Podman networking. Ensure slirp4netns is installed and properly configured."
	}
	
	if strings.Contains(errStr, "user namespace") {
		return "This may be related to user namespace mapping. Check /etc/subuid and /etc/subgid configuration."
	}
	
	return "Check Podman documentation at https://podman.io/getting-started/ for troubleshooting."
}

// isRootlessPodman detects if we're running in rootless Podman mode
func isRootlessPodman(ctx context.Context, cli client.APIClient) bool {
	if cli == nil {
		return false
	}
	
	// Try to detect rootless mode by checking system info
	info, err := cli.Info(ctx)
	if err != nil {
		return false
	}
	
	// Podman includes security options that indicate rootless mode
	for _, secOpt := range info.SecurityOptions {
		if strings.Contains(strings.ToLower(secOpt), "rootless") {
			return true
		}
	}
	
	return false
}