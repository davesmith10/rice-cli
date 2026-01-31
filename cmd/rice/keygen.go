package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/davesmith10/rice-cli/internal/sign"
	"github.com/spf13/cobra"
)

func init() {
	// Add keygen as a subcommand
}

func keygenCmd() *cobra.Command {
	var outputDir string

	cmd := &cobra.Command{
		Use:   "keygen",
		Short: "Generate Ed25519 key pair for signing",
		Long:  `Generate a new Ed25519 key pair for signing ricecake bundles.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runKeygen(outputDir)
		},
	}

	homeDir, _ := os.UserHomeDir()
	defaultDir := filepath.Join(homeDir, ".rice")

	cmd.Flags().StringVar(&outputDir, "output", defaultDir, "Output directory for keys")

	return cmd
}

func runKeygen(outputDir string) error {
	// Create output directory if it doesn't exist
	if err := os.MkdirAll(outputDir, 0700); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Check if keys already exist
	privatePath := filepath.Join(outputDir, "private.key")
	publicPath := filepath.Join(outputDir, "public.key")

	if _, err := os.Stat(privatePath); err == nil {
		return fmt.Errorf("private key already exists at %s (remove it first to generate new keys)", privatePath)
	}
	if _, err := os.Stat(publicPath); err == nil {
		return fmt.Errorf("public key already exists at %s (remove it first to generate new keys)", publicPath)
	}

	// Generate key pair
	fmt.Println("Generating Ed25519 key pair...")
	publicKey, privateKey, err := sign.GenerateKeyPair()
	if err != nil {
		return fmt.Errorf("failed to generate keys: %w", err)
	}

	// Save keys
	if err := sign.SaveKeyPair(publicKey, privateKey, outputDir); err != nil {
		return fmt.Errorf("failed to save keys: %w", err)
	}

	fmt.Println()
	fmt.Println("Key pair generated successfully!")
	fmt.Printf("  Private key: %s (keep secret!)\n", privatePath)
	fmt.Printf("  Public key:  %s (embed in player)\n", publicPath)
	fmt.Println()
	fmt.Println("To sign bundles, use:")
	fmt.Printf("  rice sign mybundle.ricecake --key %s\n", privatePath)

	return nil
}
