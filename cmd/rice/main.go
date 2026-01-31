package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	version = "1.0.0"
	verbose bool
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "rice",
		Short: "Ricecake bundle management tool",
		Long: `Rice is the administrator's toolkit for creating, validating,
signing, and testing ricecake bundles.

A ricecake is a self-contained music release bundle distributed
as a single archive file.`,
		Version: version,
	}

	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "V", false, "Enable verbose output")

	rootCmd.AddCommand(initCmd())
	rootCmd.AddCommand(buildCmd())
	rootCmd.AddCommand(validateCmd())
	rootCmd.AddCommand(signCmd())
	rootCmd.AddCommand(testCmd())
	rootCmd.AddCommand(infoCmd())
	rootCmd.AddCommand(describeCmd())
	rootCmd.AddCommand(keygenCmd())

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
