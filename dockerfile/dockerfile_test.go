
package dockerfile

import (
	"strings"
	"testing"
)

// TestGenerateDockerfileContent_FullConfig tests the GenerateDockerfileContent function with a comprehensive configuration.
func TestGenerateDockerfileContent_FullConfig(t *testing.T) {
	config := DockerfileConfig{
		BaseImage:      "ubuntu:latest",
		AppPort:        8080,
		Dependencies:   []string{"git", "curl"},
		CopyFiles:      []CopyInstruction{{Origin: "source", Destination: "dest"}},
		BuildCommand:   "go build -o myapp",
		PreRunCommands: []string{"chmod +x myapp"},
		RunCommand:     "./myapp",
		Entrypoint:     "/bin/bash",
		Workspace:      "/app",
		ExposePort:     true,
		User:           "testuser",
	}

	content, err := config.GenerateDockerfileContent()
	if err != nil {
		t.Fatalf("Expected no error, but got: %v", err)
	}

	// Check for key instructions in the generated content
	expectedStrings := []string{
		"FROM ubuntu:latest",
		"USER testuser",
		"RUN apt-get update && apt-get install -y",
		"git",
		"curl",
		"COPY source dest",
		"RUN go build -o myapp",
		"RUN chmod +x myapp",
		"EXPOSE 8080",
		"WORKDIR /app",
		`ENTRYPOINT ["/bin/bash"]`,
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(content, expected) {
			t.Errorf("Expected Dockerfile to contain '%s', but it did not.", expected)
		}
	}
}

// TestGenerateDockerfileContent_MinimalConfig tests with only the essential fields.
func TestGenerateDockerfileContent_MinimalConfig(t *testing.T) {
	config := DockerfileConfig{
		BaseImage: "alpine:latest",
	}

	content, err := config.GenerateDockerfileContent()
	if err != nil {
		t.Fatalf("Expected no error, but got: %v", err)
	}

	expected := "FROM alpine:latest"
	if !strings.Contains(content, expected) {
		t.Errorf("Expected Dockerfile to contain '%s', but it did not.", expected)
	}
}

// TestGenerateDockerfileContent_NoBaseImage tests that an error is returned when the base image is missing.
func TestGenerateDockerfileContent_NoBaseImage(t *testing.T) {
	config := DockerfileConfig{} // No base image

	_, err := config.GenerateDockerfileContent()
	if err == nil {
		t.Fatal("Expected an error when BaseImage is empty, but got none.")
	}

	expectedError := "base image cannot be empty"
	if err.Error() != expectedError {
		t.Errorf("Expected error message '%s', but got '%s'.", expectedError, err.Error())
	}
}

// TestAddDependency tests the AddDependency method.
func TestAddDependency(t *testing.T) {
	config := DockerfileConfig{BaseImage: "test"}
	config.AddDependency("new-dep")

	if len(config.Dependencies) != 1 || config.Dependencies[0] != "new-dep" {
		t.Errorf("Expected 'new-dep' to be added to dependencies, but it was not.")
	}
}

// TestAddCopyFile tests the AddCopyFile method.
func TestAddCopyFile(t *testing.T) {
	config := DockerfileConfig{BaseImage: "test"}
	newCopy := CopyInstruction{Origin: "o", Destination: "d"}
	config.AddCopyFile(newCopy)

	if len(config.CopyFiles) != 1 || config.CopyFiles[0] != newCopy {
		t.Errorf("Expected a new copy instruction to be added, but it was not.")
	}
}

// TestAddPreRunCommand tests the AddPreRunCommand method.
func TestAddPreRunCommand(t *testing.T) {
	config := DockerfileConfig{BaseImage: "test"}
	newCmd := "echo 'hello'"
	config.AddPreRunCommand(newCmd)

	if len(config.PreRunCommands) != 1 || config.PreRunCommands[0] != newCmd {
		t.Errorf("Expected a new pre-run command to be added, but it was not.")
	}
}
