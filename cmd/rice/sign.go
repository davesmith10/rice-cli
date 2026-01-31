package main

import (
	"crypto/ed25519"
	"fmt"
	"os"
	"path/filepath"

	"github.com/davesmith10/rice-cli/internal/sign"
	"github.com/spf13/cobra"
)

func signCmd() *cobra.Command {
	var keyPath, keyEnv string

	cmd := &cobra.Command{
		Use:   "sign [bundle]",
		Short: "Add digital signature to a bundle",
		Long:  `Add a digital signature to a ricecake bundle using Ed25519.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSign(args[0], keyPath, keyEnv)
		},
	}

	cmd.Flags().StringVar(&keyPath, "key", "", "Path to private key file")
	cmd.Flags().StringVar(&keyEnv, "key-env", "RICE_SIGNING_KEY", "Environment variable containing key")

	return cmd
}

func runSign(bundlePath, keyPath, keyEnv string) error {
	// Check bundle exists
	if _, err := os.Stat(bundlePath); os.IsNotExist(err) {
		return fmt.Errorf("bundle not found: %s", bundlePath)
	}

	// Load private key
	var privateKey ed25519.PrivateKey
	var err error

	if keyPath != "" {
		fmt.Printf("Loading key from: %s\n", keyPath)
		privateKey, err = sign.LoadPrivateKey(keyPath)
		if err != nil {
			return fmt.Errorf("failed to load private key: %w", err)
		}
	} else {
		fmt.Printf("Loading key from environment: %s\n", keyEnv)
		privateKey, err = sign.LoadPrivateKeyFromEnv(keyEnv)
		if err != nil {
			return fmt.Errorf("failed to load private key from environment: %w", err)
		}
	}

	fmt.Printf("Signing bundle: %s\n\n", filepath.Base(bundlePath))

	// Create signer and sign bundle
	signer := sign.NewSigner(privateKey, verbose)
	if err := signer.SignBundle(bundlePath); err != nil {
		return fmt.Errorf("signing failed: %w", err)
	}

	fmt.Println()
	fmt.Println("Bundle signed successfully.")
	fmt.Println("  Signature: signature.sig")

	return nil
}
