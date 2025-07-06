package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gdesouza/godockerfile/types"
)

func main() {
	// Command-line flags for all primitive DockerfileConfig fields
	baseImage := flag.String("base", "", "Base image for the Dockerfile (required)")
	appPort := flag.Int("port", 0, "Port to expose (optional)")
	deps := flag.String("deps", "", "Comma-separated list of dependencies (optional)")
	buildCmd := flag.String("build", "", "Build command (optional)")
	preRun := flag.String("prerun", "", "Comma-separated list of pre-run commands (optional)")
	runCmd := flag.String("run", "", "Run command (optional)")
	entrypoint := flag.String("entrypoint", "", "Entrypoint for the container (optional)")
	workspace := flag.String("workspace", "", "Workspace directory (optional)")
	exposePort := flag.Bool("expose", false, "Expose the application port (optional)")
	user := flag.String("user", "", "User to run the application as (optional)")
	outputDir := flag.String("out", ".", "Output directory for the Dockerfile (optional)")

	flag.Parse()

	// Validate required parameter
	if *baseImage == "" {
		fmt.Fprintln(os.Stderr, "Error: --base is required")
		os.Exit(1)
	}

	// Parse comma-separated lists
	var dependencies []string
	if *deps != "" {
		for _, dep := range strings.Split(*deps, ",") {
			trimmed := strings.TrimSpace(dep)
			if trimmed != "" {
				dependencies = append(dependencies, trimmed)
			}
		}
	}

	var preRunCommands []string
	if *preRun != "" {
		for _, cmd := range strings.Split(*preRun, ",") {
			trimmed := strings.TrimSpace(cmd)
			if trimmed != "" {
				preRunCommands = append(preRunCommands, trimmed)
			}
		}
	}

	// Instantiate DockerfileConfig
	config := types.DockerfileConfig{
		BaseImage:      *baseImage,
		AppPort:        *appPort,
		Dependencies:   dependencies,
		BuildCommand:   *buildCmd,
		PreRunCommands: preRunCommands,
		RunCommand:     *runCmd,
		Entrypoint:     *entrypoint,
		Workspace:      *workspace,
		ExposePort:     *exposePort,
		User:           *user,
	}

	// Generate Dockerfile content
	content, err := config.GenerateDockerfileContent()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating Dockerfile: %v\n", err)
		os.Exit(1)
	}

	// Ensure output directory exists
	if err := os.MkdirAll(*outputDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating output directory: %v\n", err)
		os.Exit(1)
	}

	// Write Dockerfile
	outPath := filepath.Join(*outputDir, "Dockerfile")
	if err := os.WriteFile(outPath, []byte(content), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing Dockerfile: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Dockerfile generated at %s\n", outPath)
}
