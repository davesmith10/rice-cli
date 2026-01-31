package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

func initCmd() *cobra.Command {
	var artist, title string
	var tracks int

	cmd := &cobra.Command{
		Use:   "init [directory]",
		Short: "Initialize a new bundle directory",
		Long:  `Initialize a new bundle directory with template files for creating a ricecake bundle.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			dir := args[0]
			return runInit(dir, artist, title, tracks)
		},
	}

	cmd.Flags().StringVar(&artist, "artist", "", "Artist or band name")
	cmd.Flags().StringVar(&title, "title", "", "Album or release title")
	cmd.Flags().IntVar(&tracks, "tracks", 1, "Number of tracks to template")

	return cmd
}

func runInit(dir, artist, title string, trackCount int) error {
	// Check if directory already exists
	if _, err := os.Stat(dir); !os.IsNotExist(err) {
		return fmt.Errorf("directory already exists: %s", dir)
	}

	// Create directory structure
	dirs := []string{
		dir,
		filepath.Join(dir, "audio"),
		filepath.Join(dir, "images"),
		filepath.Join(dir, "liner-notes"),
	}

	for _, d := range dirs {
		if err := os.MkdirAll(d, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", d, err)
		}
	}

	// Generate manifest.yaml
	if err := generateManifest(dir, artist, title, trackCount); err != nil {
		return fmt.Errorf("failed to create manifest.yaml: %w", err)
	}

	// Generate copyright.txt
	if err := generateCopyright(dir, artist, title); err != nil {
		return fmt.Errorf("failed to create copyright.txt: %w", err)
	}

	// Create README in images directory
	if err := createImagesReadme(dir); err != nil {
		return fmt.Errorf("failed to create images README: %w", err)
	}

	// Print success message
	fmt.Printf("Created bundle directory: %s/\n", dir)
	fmt.Println("  - manifest.yaml (edit this file)")
	fmt.Println("  - copyright.txt (edit this file)")
	fmt.Println("  - audio/ (add your audio files here)")
	fmt.Println("  - images/ (add cover.jpg here)")
	fmt.Println("  - liner-notes/ (optional)")
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Println("  1. Add audio files to audio/")
	fmt.Println("  2. Add cover.jpg to images/")
	fmt.Println("  3. Edit manifest.yaml with track information")
	fmt.Println("  4. Edit copyright.txt with your copyright info")
	fmt.Printf("  5. Run: rice validate %s/\n", dir)
	fmt.Printf("  6. Run: rice build %s/\n", dir)

	return nil
}

func generateManifest(dir, artist, title string, trackCount int) error {
	if artist == "" {
		artist = "Artist Name"
	}
	if title == "" {
		title = "Album Title"
	}

	now := time.Now()
	bundleID := uuid.New().String()

	manifest := fmt.Sprintf(`# Manifest Version (for future compatibility)
manifest_version: 1

# Release Information
release:
  title: "%s"
  artist: "%s"
  release_date: "%s"
  genre: "Genre"
  subgenre: ""
  catalog_number: ""

# Track Listing
tracks:
`, title, artist, now.Format("2006-01-02"))

	for i := 1; i <= trackCount; i++ {
		manifest += fmt.Sprintf(`  - number: %d
    title: "Track %d Title"
    duration: "0:00"
    filename: "%03d-track-%d-title"
    composers: ["%s"]
    performers: ["%s"]

`, i, i, i, i, artist, artist)
	}

	manifest += fmt.Sprintf(`# Available Audio Formats
audio_formats:
  - format: mp3
    bitrate: 320

  - format: flac
    bit_depth: 16
    sample_rate: 44100

# Image Assets
images:
  cover:
    filename: "cover.jpg"
    width: 1400
    height: 1400

# Copyright and Rights
rights:
  copyright_year: %d
  copyright_holder: "%s"
  license: "All Rights Reserved"
  contact: ""

# Bundle Metadata
bundle:
  created_by: "rice-cli v%s"
  created_at: "%s"
  bundle_id: "%s"
`, now.Year(), artist, version, now.Format(time.RFC3339), bundleID)

	return os.WriteFile(filepath.Join(dir, "manifest.yaml"), []byte(manifest), 0644)
}

func generateCopyright(dir, artist, title string) error {
	if artist == "" {
		artist = "[Artist Name]"
	}
	if title == "" {
		title = "[Album Title]"
	}

	now := time.Now()

	copyright := fmt.Sprintf(`Copyright Declaration for Ricecake Bundle
==========================================

Release Title: %s
Artist: %s
Copyright Year: %d

Rights Statement:
All audio recordings, artwork, and written content contained in this
bundle are protected by copyright law.

Copyright Holder: %s
Contact: [Email Address]

License: All Rights Reserved

[Additional terms or permissions if applicable]

Date: %s
Digital Signature: [Optional - Reference to signature.sig]
`, title, artist, now.Year(), artist, now.Format("2006-01-02"))

	return os.WriteFile(filepath.Join(dir, "copyright.txt"), []byte(copyright), 0644)
}

func createImagesReadme(dir string) error {
	readme := `Images Directory
================

Place your album artwork here.

Required:
  - cover.jpg (minimum 1400x1400 pixels, square)

Optional:
  - cover-large.jpg (3000x3000 recommended)
  - back.jpg
  - artist.jpg

All images must be JPEG format.
`
	return os.WriteFile(filepath.Join(dir, "images", "README.txt"), []byte(readme), 0644)
}
