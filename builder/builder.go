package builder

import (
	"fmt"
	"strings"
)

// create an enum for the docker commands RUN, COPY, ADD, etc.
type DockerCmdType string

const (
	FROM        DockerCmdType = "FROM"
	RUN         DockerCmdType = "RUN"
	COPY        DockerCmdType = "COPY"
	ADD         DockerCmdType = "ADD"
	WORKDIR     DockerCmdType = "WORKDIR"
	EXPOSE      DockerCmdType = "EXPOSE"
	USER        DockerCmdType = "USER"
	ENTRYPOINT  DockerCmdType = "ENTRYPOINT"
	CMD         DockerCmdType = "CMD"
	VOLUME      DockerCmdType = "VOLUME"
	ARG         DockerCmdType = "ARG"
	LABEL       DockerCmdType = "LABEL"
	ENV         DockerCmdType = "ENV"
	HEALTHCHECK DockerCmdType = "HEALTHCHECK"
	MAINTAINER  DockerCmdType = "MAINTAINER"
	STOP_SIGNAL DockerCmdType = "STOPSIGNAL"
)

type DockerCmd struct {
	Type    DockerCmdType // The type of Docker command (e.g., RUN, COPY)
	Command string        // The command to execute (e.g., "apt-get update && apt-get install -y curl")
	Args    []string      // Additional arguments for the command, if any
}

// CopyInstruction represents a single COPY instruction in a Dockerfile,
// specifying the source path (Origin) and the destination path (Destination)
// within the Docker image.
type CopyInstruction struct {
	Origin      string
	Destination string
}

// DockerfileConfig holds the configuration parameters for generating the Dockerfile.
type DockerfileConfig struct {
	BaseImage         string
	AppPort           int
	Dependencies      []string          // e.g., "git", "curl"
	CopyFiles         []CopyInstruction // e.g., "main.go", "go.mod", "go.sum"
	BuildCommand      string            // e.g., "go build -o app ."
	PreRunCommands    []string          // List of commands to run before the main CMD, e.g., "chmod +x ./app"
	RunCommand        string            // e.g., "./app"
	Entrypoint        string            // e.g., "/bin/sh -c"
	Workspace         string            // Directory where the application will run, e.g., "/app"
	ExposePort        bool
	User              string      // New: User to run the application as, e.g., "nonroot" or "appuser"
	OrderedDockerCmds []DockerCmd // List of generic Docker commands to include in the Dockerfile
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

// AddOrderedDockerCmd appends a new generic Docker command to the OrderedDockerCmds list.
func (config *DockerfileConfig) AddOrderedDockerCmd(cmd DockerCmd) {
	config.OrderedDockerCmds = append(config.OrderedDockerCmds, cmd)
}

// GenerateDockerfileContent creates the content of a Dockerfile as a string
// based on the provided DockerfileConfig.
func (config *DockerfileConfig) GenerateDockerfileContent() (string, error) {
	if config.BaseImage == "" {
		return "", fmt.Errorf("base image cannot be empty")
	}

	var builder strings.Builder

	// add header information
	builder.WriteString("# Auto-generated Dockerfile\n")
	builder.WriteString("# Do not edit this file manually\n")

	// Add version information from builder/version.go
	builder.WriteString(fmt.Sprintf("# Dockerbot version: %s\n", Version))

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

	// Add ordered Docker commands if any
	if len(config.OrderedDockerCmds) > 0 {
		for _, cmd := range config.OrderedDockerCmds {
			line, err := formatDockerCmd(cmd)
			if err != nil {
				return "", err
			}
			builder.WriteString(line)
		}
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

	// Set working directory if provided
	if config.Workspace != "" {
		builder.WriteString(fmt.Sprintf("WORKDIR %s\n", config.Workspace))
		builder.WriteString("\n")
	}

	// Define the entrypoint/command to run the application
	if config.Entrypoint != "" {
		builder.WriteString(fmt.Sprintf("ENTRYPOINT [\"%s\"]\n", config.Entrypoint))
	}

	return builder.String(), nil
}

func formatDockerCmd(cmd DockerCmd) (string, error) {
	switch cmd.Type {
	case RUN, WORKDIR, USER, MAINTAINER:
		return fmt.Sprintf("%s %s\n", cmd.Type, cmd.Command), nil
	case COPY, ADD:
		if len(cmd.Args) > 0 {
			return fmt.Sprintf("%s %s %s\n", cmd.Type, cmd.Command, strings.Join(cmd.Args, " ")), nil
		}
		return fmt.Sprintf("%s %s\n", cmd.Type, cmd.Command), nil
	case EXPOSE, STOP_SIGNAL:
		if len(cmd.Args) > 0 {
			return fmt.Sprintf("%s %s\n", cmd.Type, strings.Join(cmd.Args, " ")), nil
		}
		return fmt.Sprintf("%s %s\n", cmd.Type, cmd.Command), nil
	case ENTRYPOINT, CMD:
		if len(cmd.Args) > 0 {
			return fmt.Sprintf("%s [\"%s\", \"%s\"]\n", cmd.Type, cmd.Command, strings.Join(cmd.Args, "\", \"")), nil
		}
		return fmt.Sprintf("%s [\"%s\"]\n", cmd.Type, cmd.Command), nil
	case VOLUME:
		if len(cmd.Args) > 0 {
			return fmt.Sprintf("VOLUME [\"%s\"]\n", strings.Join(cmd.Args, "\", \"")), nil
		}
		return fmt.Sprintf("VOLUME [\"%s\"]\n", cmd.Command), nil
	case ARG, LABEL, ENV:
		if len(cmd.Args) > 0 {
			return fmt.Sprintf("%s %s=%s\n", cmd.Type, cmd.Command, strings.Join(cmd.Args, " ")), nil
		}
		return fmt.Sprintf("%s %s\n", cmd.Type, cmd.Command), nil
	case HEALTHCHECK:
		if len(cmd.Args) > 0 {
			return fmt.Sprintf("HEALTHCHECK CMD %s\n", strings.Join(cmd.Args, " ")), nil
		}
		return fmt.Sprintf("HEALTHCHECK CMD %s\n", cmd.Command), nil
	default:
		return "", fmt.Errorf("unknown Docker command type: %s", cmd.Type)
	}
}
