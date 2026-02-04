package container

import (
	"os"
	"strings"
	"testing"
)

func TestNewRuntimeDetector(t *testing.T) {
	detector := NewRuntimeDetector()
	if detector == nil {
		t.Fatal("NewRuntimeDetector() returned nil")
	}
}

func TestContainerRuntimeString(t *testing.T) {
	tests := []struct {
		runtime ContainerRuntime
		want    string
	}{
		{RuntimeDocker, "docker"},
		{RuntimePodman, "podman"},
		{RuntimeUnknown, "unknown"},
	}

	for _, tt := range tests {
		if got := tt.runtime.String(); got != tt.want {
			t.Errorf("ContainerRuntime.String() = %v, want %v", got, tt.want)
		}
	}
}

func TestSetPreferredRuntime(t *testing.T) {
	detector := NewRuntimeDetector()
	
	detector.SetPreferredRuntime(RuntimeDocker)
	if detector.preferredRuntime != RuntimeDocker {
		t.Errorf("SetPreferredRuntime(RuntimeDocker) failed")
	}
	
	detector.SetPreferredRuntime(RuntimePodman)
	if detector.preferredRuntime != RuntimePodman {
		t.Errorf("SetPreferredRuntime(RuntimePodman) failed")
	}
}

func TestSetCustomSocket(t *testing.T) {
	detector := NewRuntimeDetector()
	testSocket := "/test/socket"
	
	detector.SetCustomSocket(testSocket)
	if detector.customSocket != testSocket {
		t.Errorf("SetCustomSocket(%s) failed", testSocket)
	}
}

func TestCheckEnvironmentHints(t *testing.T) {
	detector := NewRuntimeDetector()
	
	// Test ACT_CONTAINER_RUNTIME environment variable
	tests := []struct {
		envVar string
		value  string
		want   ContainerRuntime
	}{
		{"ACT_CONTAINER_RUNTIME", "docker", RuntimeDocker},
		{"ACT_CONTAINER_RUNTIME", "podman", RuntimePodman},
		{"ACT_CONTAINER_RUNTIME", "DOCKER", RuntimeDocker}, // case insensitive
		{"ACT_CONTAINER_RUNTIME", "PODMAN", RuntimePodman}, // case insensitive
		{"ACT_CONTAINER_RUNTIME", "invalid", RuntimeUnknown},
		{"PODMAN_HOST", "unix:///test", RuntimePodman},
		{"DOCKER_HOST", "unix:///test", RuntimeDocker},
	}
	
	for _, tt := range tests {
		// Clear all relevant env vars first
		os.Unsetenv("ACT_CONTAINER_RUNTIME")
		os.Unsetenv("PODMAN_HOST")
		os.Unsetenv("DOCKER_HOST")
		
		// Set the test env var
		os.Setenv(tt.envVar, tt.value)
		
		got := detector.checkEnvironmentHints()
		if got != tt.want {
			t.Errorf("checkEnvironmentHints() with %s=%s = %v, want %v", tt.envVar, tt.value, got, tt.want)
		}
		
		// Clean up
		os.Unsetenv(tt.envVar)
	}
}

func TestGuessRuntimeFromSocket(t *testing.T) {
	detector := NewRuntimeDetector()
	
	tests := []struct {
		socket string
		want   ContainerRuntime
	}{
		{"/run/podman/podman.sock", RuntimePodman},
		{"/var/run/docker.sock", RuntimeDocker},
		{"unix:///run/podman/podman.sock", RuntimePodman},
		{"unix:///var/run/docker.sock", RuntimeDocker},
		{"/some/unknown/socket.sock", RuntimeDocker}, // defaults to docker
	}
	
	for _, tt := range tests {
		got := detector.guessRuntimeFromSocket(tt.socket)
		if got != tt.want {
			t.Errorf("guessRuntimeFromSocket(%s) = %v, want %v", tt.socket, got, tt.want)
		}
	}
}

func TestGetHelpfulErrorMessage(t *testing.T) {
	detector := NewRuntimeDetector()
	
	message := detector.GetHelpfulErrorMessage()
	
	// Verify the message contains expected elements
	expectedElements := []string{
		"No container runtime detected",
		"Docker",
		"Podman",
		"act --container-runtime",
		"act --container-socket",
	}
	
	for _, element := range expectedElements {
		if !strings.Contains(message, element) {
			t.Errorf("GetHelpfulErrorMessage() missing expected element: %s", element)
		}
	}
}

func TestRuntimeDetectionPriority(t *testing.T) {
	detector := NewRuntimeDetector()
	
	// Test that preferred runtime takes precedence
	detector.SetPreferredRuntime(RuntimeDocker)
	
	// This would normally trigger auto-detection, but preferred should win
	// Since we can't easily test actual runtime availability in unit tests,
	// we'll test the logic flow
	
	// The actual detection depends on binary/socket availability
	// so we mainly test the configuration aspects here
	if detector.preferredRuntime != RuntimeDocker {
		t.Errorf("Preferred runtime not set correctly")
	}
}

func TestFactoryIntegration(t *testing.T) {
	// Test the factory integration
	input := &NewContainerInput{
		Image: "test:latest",
		Name:  "test-container",
	}
	
	// This should not crash and should return some implementation
	container := NewContainer(input)
	if container == nil {
		t.Fatal("NewContainer() returned nil")
	}
	
	// Test with override
	SetRuntimeOverride(RuntimeDocker)
	defer ClearRuntimeOverride()
	
	container = NewContainer(input)
	if container == nil {
		t.Fatal("NewContainer() with override returned nil")
	}
}

func TestGetCurrentRuntime(t *testing.T) {
	// Test getting current runtime
	runtime := GetCurrentRuntime()
	
	// Should return some valid runtime (even if unknown/stub)
	validRuntimes := []ContainerRuntime{RuntimeUnknown, RuntimeDocker, RuntimePodman}
	found := false
	for _, valid := range validRuntimes {
		if runtime == valid {
			found = true
			break
		}
	}
	
	if !found {
		t.Errorf("GetCurrentRuntime() returned invalid runtime: %v", runtime)
	}
}

func TestGetAvailableRuntimes(t *testing.T) {
	// Test getting available runtimes
	runtimes := GetAvailableRuntimes()
	
	// Should return a slice (may be empty if no runtimes available)
	if runtimes == nil {
		t.Fatal("GetAvailableRuntimes() returned nil")
	}
	
	// Each runtime in the slice should be valid
	for _, runtime := range runtimes {
		if runtime != RuntimeDocker && runtime != RuntimePodman {
			t.Errorf("GetAvailableRuntimes() returned invalid runtime: %v", runtime)
		}
	}
}