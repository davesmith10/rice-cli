package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/davesmith10/rice-cli/internal/bundle"
	"github.com/davesmith10/rice-cli/internal/validate"
	"github.com/spf13/cobra"
)

func buildCmd() *cobra.Command {
	var output string
	var noValidate, force bool

	cmd := &cobra.Command{
		Use:   "build [directory]",
		Short: "Create a bundle from directory contents",
		Long:  `Create a ricecake bundle from a directory containing music and metadata.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runBuild(args[0], output, noValidate, force)
		},
	}

	cmd.Flags().StringVar(&output, "output", "", "Output bundle path (default: [directory].ricecake)")
	cmd.Flags().BoolVar(&noValidate, "no-validate", false, "Skip validation before building")
	cmd.Flags().BoolVar(&force, "force", false, "Overwrite existing bundle")

	return cmd
}

func runBuild(dir, output string, noValidate, force bool) error {
	// Clean up directory path
	dir = strings.TrimSuffix(dir, "/")
	dir = strings.TrimSuffix(dir, "\\")

	// Check source directory exists
	info, err := os.Stat(dir)
	if err != nil {
		return fmt.Errorf("source directory not found: %s", dir)
	}
	if !info.IsDir() {
		return fmt.Errorf("not a directory: %s", dir)
	}

	// Determine output path
	if output == "" {
		output = dir + ".ricecake"
	}

	// Check if output already exists
	if _, err := os.Stat(output); err == nil {
		if !force {
			return fmt.Errorf("output file already exists: %s (use --force to overwrite)", output)
		}
		os.Remove(output)
	}

	// Run validation unless skipped
	if !noValidate {
		fmt.Println("Validating bundle...")
		validator := validate.New(dir, false)
		report, err := validator.Validate()
		if err != nil {
			return fmt.Errorf("validation failed: %w", err)
		}

		if report.Errors > 0 {
			// Print errors
			for _, result := range report.Results {
				if !result.Passed && result.Severity == "error" {
					fmt.Printf("  [ERROR] %s: %s\n", result.Check, result.Message)
				}
			}
			return fmt.Errorf("validation failed with %d error(s)", report.Errors)
		}

		fmt.Println("  Validation passed")
		fmt.Println()
	}

	// Build the bundle
	fmt.Println("Building bundle...")
	builder := bundle.NewBuilder(dir, output, verbose)
	if err := builder.Build(); err != nil {
		return fmt.Errorf("build failed: %w", err)
	}

	// Get bundle info
	bundleInfo, err := bundle.GetBundleInfo(output)
	if err != nil {
		return fmt.Errorf("failed to get bundle info: %w", err)
	}

	fmt.Println()
	fmt.Printf("Bundle created: %s\n", filepath.Base(output))
	fmt.Printf("  Size: %s\n", bundle.FormatSize(bundleInfo.Size))

	return nil
}
