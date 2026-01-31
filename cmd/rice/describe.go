package main

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

func describeCmd() *cobra.Command {
	var rawOutput bool

	cmd := &cobra.Command{
		Use:   "describe [bundle-or-directory]",
		Short: "Print the contents of a ricecake bundle",
		Long: `Print the manifest.yaml contents of a ricecake bundle.

This command reads and displays the manifest file from either a source
directory or a .ricecake bundle file.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDescribe(args[0], rawOutput)
		},
	}

	cmd.Flags().BoolVar(&rawOutput, "raw", false, "Output raw YAML without any formatting")

	return cmd
}

func runDescribe(path string, rawOutput bool) error {
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("path not found: %s", path)
	}

	var manifestData []byte

	if info.IsDir() {
		// Read manifest from directory
		manifestPath := filepath.Join(path, "manifest.yaml")
		manifestData, err = os.ReadFile(manifestPath)
		if err != nil {
			return fmt.Errorf("failed to read manifest: %w", err)
		}
	} else {
		// Read manifest from .ricecake bundle (ZIP file)
		manifestData, err = readManifestFromBundle(path)
		if err != nil {
			return fmt.Errorf("failed to read manifest from bundle: %w", err)
		}
	}

	if !rawOutput {
		fmt.Printf("# Manifest from: %s\n", path)
		fmt.Println("# " + string(make([]byte, 0)) + "---")
	}

	fmt.Print(string(manifestData))

	// Ensure there's a trailing newline
	if len(manifestData) > 0 && manifestData[len(manifestData)-1] != '\n' {
		fmt.Println()
	}

	return nil
}

// readManifestFromBundle reads the manifest.yaml from a .ricecake ZIP archive
func readManifestFromBundle(bundlePath string) ([]byte, error) {
	reader, err := zip.OpenReader(bundlePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open bundle: %w", err)
	}
	defer reader.Close()

	// Look for manifest.yaml in the archive
	for _, file := range reader.File {
		if file.Name == "manifest.yaml" || file.Name == "./manifest.yaml" {
			rc, err := file.Open()
			if err != nil {
				return nil, fmt.Errorf("failed to open manifest in bundle: %w", err)
			}
			defer rc.Close()

			data, err := io.ReadAll(rc)
			if err != nil {
				return nil, fmt.Errorf("failed to read manifest from bundle: %w", err)
			}

			return data, nil
		}
	}

	return nil, fmt.Errorf("manifest.yaml not found in bundle")
}
