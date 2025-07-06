package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/gdesouza/godockerfile/types"
)

func main() {
	// example usage of the Dockerfile generation function
	rootRepoPath := "/path/to/your/root/repo" // Replace with your actual root repo path
	dockerfileName := "Dockerfile"
	outputFilePath := filepath.Join(rootRepoPath, dockerfileName)

	dockerfileConfig := types.DockerfileConfig{
		BaseImage:      "ubuntu:20.04",
		AppPort:        8080,
		Dependencies:   []string{"git", "curl"},
		CopyFiles:      []types.CopyInstruction{{Origin: "main.go", Destination: "/app/main.go"}},
		BuildCommand:   "go build -o app .",
		PreRunCommands: []string{"chmod +x ./app"},
		RunCommand:     "./app",
		Entrypoint:     "/bin/sh -c",
		Workspace:      "/app",
		ExposePort:     true,
		User:           "nonroot", // Example user
	}

	// add dependencies and copy files
	dockerfileConfig.AddDependency("ros-noetic-ros-base")
	dockerfileConfig.AddCopyFile(types.CopyInstruction{
		Origin:      "config.file",
		Destination: "config.file",
	})

	// add pre-run commands
	dockerfileConfig.AddPreRunCommand("groupadd --gid 1000 mygroup")
	dockerfileConfig.AddPreRunCommand("useradd --uid 1000 --gid 1000 myuser")

	// generate the Dockerfile content
	dockerfileContent, err := dockerfileConfig.GenerateDockerfileContent()
	if err != nil {
		fmt.Printf("Error generating Dockerfile content: %v\n", err)
		return
	}

	// write the Dockerfile content to the output file
	if err := os.WriteFile(outputFilePath, []byte(dockerfileContent), 0644); err != nil {
		fmt.Printf("Error writing Dockerfile to %s: %v\n", outputFilePath, err)
		return
	}

	// print success message
	fmt.Printf("Dockerfile generated successfully at %s\n", outputFilePath)
}
