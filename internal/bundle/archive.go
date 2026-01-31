package bundle

import (
	"archive/zip"
	"compress/flate"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// Builder creates ricecake bundles
type Builder struct {
	sourceDir  string
	outputPath string
	verbose    bool
}

// NewBuilder creates a new bundle builder
func NewBuilder(sourceDir, outputPath string, verbose bool) *Builder {
	return &Builder{
		sourceDir:  sourceDir,
		outputPath: outputPath,
		verbose:    verbose,
	}
}

// Build creates the ricecake bundle
func (b *Builder) Build() error {
	// Create output file
	outFile, err := os.Create(b.outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer outFile.Close()

	// Create ZIP writer with compression level 6
	zipWriter := zip.NewWriter(outFile)
	zipWriter.RegisterCompressor(zip.Deflate, func(out io.Writer) (io.WriteCloser, error) {
		return flate.NewWriter(out, 6)
	})
	defer zipWriter.Close()

	// Walk the source directory and add files
	fileCount := 0
	err = filepath.Walk(b.sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Get relative path
		relPath, err := filepath.Rel(b.sourceDir, path)
		if err != nil {
			return err
		}

		// Skip the root directory
		if relPath == "." {
			return nil
		}

		// Skip README.txt files in subdirectories (they're just helpers)
		if info.Name() == "README.txt" && filepath.Dir(relPath) != "." {
			return nil
		}

		// Convert to forward slashes for ZIP compatibility
		zipPath := strings.ReplaceAll(relPath, string(filepath.Separator), "/")

		if info.IsDir() {
			// Add directory entry
			_, err = zipWriter.Create(zipPath + "/")
			return err
		}

		// Add file
		if b.verbose {
			fmt.Printf("  Adding %s\n", relPath)
		}

		// Create ZIP entry with proper header
		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}
		header.Name = zipPath
		header.Method = zip.Deflate

		writer, err := zipWriter.CreateHeader(header)
		if err != nil {
			return err
		}

		// Copy file contents
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		_, err = io.Copy(writer, file)
		if err != nil {
			return err
		}

		fileCount++
		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to add files to bundle: %w", err)
	}

	return nil
}

// GetBundleInfo returns information about a bundle
func GetBundleInfo(bundlePath string) (*BundleInfo, error) {
	info, err := os.Stat(bundlePath)
	if err != nil {
		return nil, fmt.Errorf("cannot stat bundle: %w", err)
	}

	return &BundleInfo{
		Path: bundlePath,
		Size: info.Size(),
	}, nil
}

// BundleInfo contains bundle metadata
type BundleInfo struct {
	Path string
	Size int64
}

// FormatSize returns a human-readable size string
func FormatSize(bytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)

	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.2f GB", float64(bytes)/float64(GB))
	case bytes >= MB:
		return fmt.Sprintf("%.2f MB", float64(bytes)/float64(MB))
	case bytes >= KB:
		return fmt.Sprintf("%.2f KB", float64(bytes)/float64(KB))
	default:
		return fmt.Sprintf("%d bytes", bytes)
	}
}
