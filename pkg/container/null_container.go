package container

import (
	"context"
	"fmt"
	"io"

	"github.com/nektos/act/pkg/common"
)

// nullContainer implements the Null Object Pattern, providing helpful error messages
// when no container runtime is available instead of returning nil
type nullContainer struct {
	input *NewContainerInput
}

// newNullContainer creates a null container that provides helpful error messages
// when no container runtime is available
func newNullContainer(input *NewContainerInput) ExecutionsEnvironment {
	return &nullContainer{
		input: input,
	}
}

func (n *nullContainer) Create(capAdd []string, capDrop []string) common.Executor {
	return func(ctx context.Context) error {
		return fmt.Errorf("no container runtime available\n\n%s", GetRuntimeDetectionError())
	}
}

func (n *nullContainer) Copy(destPath string, files ...*FileEntry) common.Executor {
	return func(ctx context.Context) error {
		return fmt.Errorf("no container runtime available\n\n%s", GetRuntimeDetectionError())
	}
}

func (n *nullContainer) CopyTarStream(ctx context.Context, destPath string, tarStream io.Reader) error {
	return fmt.Errorf("no container runtime available\n\n%s", GetRuntimeDetectionError())
}

func (n *nullContainer) CopyDir(destPath string, srcPath string, useGitIgnore bool) common.Executor {
	return func(ctx context.Context) error {
		return fmt.Errorf("no container runtime available\n\n%s", GetRuntimeDetectionError())
	}
}

func (n *nullContainer) GetContainerArchive(ctx context.Context, srcPath string) (io.ReadCloser, error) {
	return nil, fmt.Errorf("no container runtime available\n\n%s", GetRuntimeDetectionError())
}

func (n *nullContainer) Pull(forcePull bool) common.Executor {
	return func(ctx context.Context) error {
		return fmt.Errorf("no container runtime available\n\n%s", GetRuntimeDetectionError())
	}
}

func (n *nullContainer) Start(attach bool) common.Executor {
	return func(ctx context.Context) error {
		return fmt.Errorf("no container runtime available\n\n%s", GetRuntimeDetectionError())
	}
}

func (n *nullContainer) Exec(command []string, env map[string]string, user, workdir string) common.Executor {
	return func(ctx context.Context) error {
		return fmt.Errorf("no container runtime available\n\n%s", GetRuntimeDetectionError())
	}
}

func (n *nullContainer) UpdateFromEnv(srcPath string, env *map[string]string) common.Executor {
	return func(ctx context.Context) error {
		return fmt.Errorf("no container runtime available\n\n%s", GetRuntimeDetectionError())
	}
}

func (n *nullContainer) UpdateFromImageEnv(env *map[string]string) common.Executor {
	return func(ctx context.Context) error {
		return fmt.Errorf("no container runtime available\n\n%s", GetRuntimeDetectionError())
	}
}

func (n *nullContainer) Remove() common.Executor {
	return func(ctx context.Context) error {
		return fmt.Errorf("no container runtime available\n\n%s", GetRuntimeDetectionError())
	}
}

func (n *nullContainer) Close() common.Executor {
	return func(ctx context.Context) error {
		return fmt.Errorf("no container runtime available\n\n%s", GetRuntimeDetectionError())
	}
}

func (n *nullContainer) ReplaceLogWriter(stdout io.Writer, stderr io.Writer) (io.Writer, io.Writer) {
	return stdout, stderr // Just return the original writers since we can't do anything
}

func (n *nullContainer) GetHealth(ctx context.Context) Health {
	return HealthUnHealthy // Always unhealthy since no runtime is available
}

// ExecutionsEnvironment interface methods
func (n *nullContainer) ToContainerPath(path string) string {
	return path // Just return the path unchanged
}

func (n *nullContainer) GetActPath() string {
	return "/opt/act" // Return a default path
}

func (n *nullContainer) GetPathVariableName() string {
	return "PATH"
}

func (n *nullContainer) DefaultPathVariable() string {
	return "/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"
}

func (n *nullContainer) JoinPathVariable(paths ...string) string {
	// Use the Linux container extensions logic
	return (&LinuxContainerEnvironmentExtensions{}).JoinPathVariable(paths...)
}

func (n *nullContainer) GetRunnerContext(ctx context.Context) map[string]interface{} {
	return map[string]interface{}{
		"os":           "linux",
		"arch":         "x64",
		"temp":         "/tmp",
		"tool_cache":   "/opt/hostedtoolcache",
		"action_path":  "/github/workspace",
		"workspace":    "/github/workspace",
	}
}

func (n *nullContainer) IsEnvironmentCaseInsensitive() bool {
	return false // Linux-style environment
}