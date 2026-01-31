package convert

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
)

// LameOptions contains options for LAME encoding
type LameOptions struct {
	Bitrate int // CBR bitrate (128, 192, 256, 320)
	Quality int // VBR quality (0-9), -1 means use CBR
}

// LameRunner manages the embedded LAME binary
type LameRunner struct {
	binaryPath string
	tempDir    string
}

// NewLameRunner extracts the embedded LAME binary and returns a runner
func NewLameRunner() (*LameRunner, error) {
	// Create temp directory for binary
	tempDir, err := os.MkdirTemp("", "rice-lame-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}

	// Determine binary name based on platform
	binaryName := "lame"
	if runtime.GOOS == "windows" {
		binaryName = "lame.exe"
	}

	binaryPath := filepath.Join(tempDir, binaryName)

	// Write embedded binary to temp file
	if err := os.WriteFile(binaryPath, lameBinary, 0755); err != nil {
		os.RemoveAll(tempDir)
		return nil, fmt.Errorf("failed to extract LAME binary: %w", err)
	}

	return &LameRunner{
		binaryPath: binaryPath,
		tempDir:    tempDir,
	}, nil
}

// Cleanup removes the temporary LAME binary
func (l *LameRunner) Cleanup() {
	if l.tempDir != "" {
		os.RemoveAll(l.tempDir)
	}
}

// Convert converts a WAV file to MP3 using LAME
func (l *LameRunner) Convert(input, output string, opts LameOptions) error {
	args := l.buildArgs(input, output, opts)

	cmd := exec.Command(l.binaryPath, args...)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		errMsg := stderr.String()
		if errMsg != "" {
			return fmt.Errorf("LAME error: %s", errMsg)
		}
		return fmt.Errorf("LAME execution failed: %w", err)
	}

	return nil
}

// buildArgs constructs the LAME command-line arguments
func (l *LameRunner) buildArgs(input, output string, opts LameOptions) []string {
	args := []string{
		"--quiet", // Suppress progress output
	}

	if opts.Quality >= 0 {
		// VBR mode
		args = append(args, "-V", strconv.Itoa(opts.Quality))
	} else {
		// CBR mode
		args = append(args, "--cbr", "-b", strconv.Itoa(opts.Bitrate))
	}

	// Input and output files
	args = append(args, input, output)

	return args
}

// GetVersion returns the version of the embedded LAME binary
func (l *LameRunner) GetVersion() (string, error) {
	cmd := exec.Command(l.binaryPath, "--version")

	var stdout bytes.Buffer
	cmd.Stdout = &stdout

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to get LAME version: %w", err)
	}

	return stdout.String(), nil
}
