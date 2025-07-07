package builder

import (
	"fmt"
	"strings"
)

// CopyInstruction represents a single COPY instruction in a Dockerfile,
// specifying the source path (Origin) and the destination path (Destination)
// within the Docker image.
type CopyInstruction struct {
	Origin      string
	Destination string
}

// DockerfileConfig holds the configuration parameters for generating the Dockerfile.
type DockerfileConfig struct {
	BaseImage      string
	AppPort        int
	Dependencies   []string          // e.g., "git", "curl"
	CopyFiles      []CopyInstruction // e.g., "main.go", "go.mod", "go.sum"
	BuildCommand   string            // e.g., "go build -o app ."
	PreRunCommands []string          // List of commands to run before the main CMD, e.g., "chmod +x ./app"
	RunCommand     string            // e.g., "./app"
	Entrypoint     string            // e.g., "/bin/sh -c"
	Workspace      string            // Directory where the application will run, e.g., "/app"
	ExposePort     bool
	User           string // New: User to run the application as, e.g., "nonroot" or "appuser"
}

// AddPreRunCommand appends a new command to the PreRunCommands list.
func (config *DockerfileConfig) AddPreRunCommand(command string) {
	config.PreRunCommands = append(config.PreRunCommands, command)
}

// AddCopyFile appends a new file to the CopyFiles list.
func (config *DockerfileConfig) AddCopyFile(files CopyInstruction) {
	config.CopyFiles = append(config.CopyFiles, files)
}

// AddDependency appends a new dependency to the Dependencies list.
func (config *DockerfileConfig) AddDependency(dependency string) {
	config.Dependencies = append(config.Dependencies, dependency)
}

// GenerateDockerfileContent creates the content of a Dockerfile as a string
// based on the provided DockerfileConfig.
func (config *DockerfileConfig) GenerateDockerfileContent() (string, error) {
	if config.BaseImage == "" {
		return "", fmt.Errorf("base image cannot be empty")
	}

	var builder strings.Builder

	// Add base image
	builder.WriteString(fmt.Sprintf("FROM %s\n", config.BaseImage))
	builder.WriteString("\n") // Add a newline for readability

	// Install dependencies if any
	if len(config.Dependencies) > 0 {
		builder.WriteString("RUN apt-get update && apt-get install -y \\\n")
		for _, dep := range config.Dependencies {
			builder.WriteString(fmt.Sprintf("    %s \\ \n", dep))
		}
		builder.WriteString("    && apt-get clean && rm -rf /var/lib/apt/lists/*\n")
		builder.WriteString("\n")
	}

	// Run pre-commands if any
	if len(config.PreRunCommands) > 0 {
		for _, cmd := range config.PreRunCommands {
			builder.WriteString(fmt.Sprintf("RUN %s\n", cmd))
		}
		builder.WriteString("\n")
	}

	// Set user if provided
	if config.User != "" {
		builder.WriteString(fmt.Sprintf("USER %s\n", config.User))
		builder.WriteString("\n")
	}

	// Copy application files
	if len(config.CopyFiles) > 0 {
		for _, file := range config.CopyFiles {
			builder.WriteString(fmt.Sprintf("COPY %s %s\n", file.Origin, file.Destination))
		}
		builder.WriteString("\n")
	} else {
		// Default copy if no specific files are provided
		builder.WriteString("COPY . .\n")
		builder.WriteString("\n")
	}

	// Build command
	if config.BuildCommand != "" {
		builder.WriteString(fmt.Sprintf("RUN %s\n", config.BuildCommand))
		builder.WriteString("\n")
	}

	// Expose port if requested
	if config.ExposePort && config.AppPort > 0 {
		builder.WriteString(fmt.Sprintf("EXPOSE %d\n", config.AppPort))
		builder.WriteString("\n")
	}

	// Set working directory
	builder.WriteString(fmt.Sprintf("WORKDIR %s\n", config.Workspace))
	builder.WriteString("\n")

	// Define the entrypoint/command to run the application
	if config.Entrypoint != "" {
		builder.WriteString(fmt.Sprintf("ENTRYPOINT [\"%s\"]\n", config.Entrypoint))
	}

	return builder.String(), nil
}
