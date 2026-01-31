package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"

	"github.com/davesmith10/rice-cli/internal/server"
	"github.com/spf13/cobra"
)

func testCmd() *cobra.Command {
	var port int
	var openBrowser bool
	var playerPath string

	cmd := &cobra.Command{
		Use:   "test [bundle-or-directory]",
		Short: "Start local preview server",
		Long:  `Start a local HTTP server to preview and test a bundle.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTest(args[0], port, openBrowser, playerPath)
		},
	}

	cmd.Flags().IntVar(&port, "port", 8080, "Server port")
	cmd.Flags().BoolVar(&openBrowser, "open", false, "Open browser automatically")
	cmd.Flags().StringVar(&playerPath, "player", "", "Path to player executable for launch")

	return cmd
}

func runTest(path string, port int, openBrowser bool, playerPath string) error {
	// Check path exists
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("path not found: %s", path)
	}

	// If it's a .ricecake file, we would need to extract it first
	// For now, only support directories
	if !info.IsDir() {
		return fmt.Errorf("currently only directories are supported (not .ricecake files)")
	}

	fmt.Println("Starting preview server...")
	fmt.Println()
	fmt.Printf("Bundle: %s\n", path)

	// Open browser if requested
	if openBrowser {
		go openURL(fmt.Sprintf("http://localhost:%d", port))
	}

	// Start preview server
	srv := server.NewPreviewServer(path, port, verbose)
	return srv.Start()
}

func openURL(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	cmd.Run()
}
