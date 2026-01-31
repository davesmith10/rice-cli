package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/davesmith10/rice-cli/internal/bundle"
	"github.com/davesmith10/rice-cli/pkg/manifest"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

func infoCmd() *cobra.Command {
	var jsonOutput, showTracks, verify bool

	cmd := &cobra.Command{
		Use:   "info [bundle]",
		Short: "Display bundle information",
		Long:  `Display detailed information about a ricecake bundle.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInfo(args[0], jsonOutput, showTracks, verify)
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")
	cmd.Flags().BoolVar(&showTracks, "tracks", false, "Show detailed track listing")
	cmd.Flags().BoolVar(&verify, "verify", false, "Verify signature if present")

	return cmd
}

func runInfo(path string, jsonOutput, showTracks, verify bool) error {
	// Determine if it's a directory or bundle file
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("path not found: %s", path)
	}

	var manifestPath string
	var bundleSize int64

	if info.IsDir() {
		manifestPath = filepath.Join(path, "manifest.yaml")
		// Calculate directory size
		filepath.Walk(path, func(_ string, info os.FileInfo, _ error) error {
			if !info.IsDir() {
				bundleSize += info.Size()
			}
			return nil
		})
	} else {
		// It's a bundle file - for now just report we don't support it yet
		return fmt.Errorf("reading .ricecake files not yet implemented - use the source directory")
	}

	// Read manifest
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return fmt.Errorf("failed to read manifest: %w", err)
	}

	var m manifest.Manifest
	if err := yaml.Unmarshal(data, &m); err != nil {
		return fmt.Errorf("failed to parse manifest: %w", err)
	}

	if jsonOutput {
		return outputInfoJSON(m, bundleSize, path)
	}

	return outputInfoText(m, bundleSize, path, showTracks, verify)
}

func outputInfoJSON(m manifest.Manifest, bundleSize int64, path string) error {
	output := map[string]interface{}{
		"path":     path,
		"manifest": m,
		"size":     bundleSize,
	}

	data, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return err
	}

	fmt.Println(string(data))
	return nil
}

func outputInfoText(m manifest.Manifest, bundleSize int64, path string, showTracks, verify bool) error {
	fmt.Println("Bundle Information")
	fmt.Println("==================")
	fmt.Println()
	fmt.Printf("Title:    %s\n", m.Release.Title)
	fmt.Printf("Artist:   %s\n", m.Release.Artist)
	fmt.Printf("Released: %s\n", m.Release.ReleaseDate)

	if m.Release.Subgenre != "" {
		fmt.Printf("Genre:    %s / %s\n", m.Release.Genre, m.Release.Subgenre)
	} else if m.Release.Genre != "" {
		fmt.Printf("Genre:    %s\n", m.Release.Genre)
	}

	fmt.Println()
	fmt.Printf("Tracks: %d\n", len(m.Tracks))

	// Calculate total duration
	totalDuration := calculateTotalDuration(m.Tracks)
	if totalDuration != "" {
		fmt.Printf("Total Duration: %s\n", totalDuration)
	}

	fmt.Println()
	fmt.Println("Audio Formats:")
	for _, af := range m.AudioFormats {
		if af.Bitrate > 0 {
			fmt.Printf("  - %s (%d kbps)\n", strings.ToUpper(af.Format), af.Bitrate)
		} else if af.BitDepth > 0 {
			fmt.Printf("  - %s (%d-bit/%dHz)\n", strings.ToUpper(af.Format), af.BitDepth, af.SampleRate)
		} else {
			fmt.Printf("  - %s\n", strings.ToUpper(af.Format))
		}
	}

	fmt.Println()
	fmt.Println("Bundle Details:")
	fmt.Printf("  Size: %s\n", bundle.FormatSize(bundleSize))
	fmt.Printf("  Created: %s\n", m.Bundle.CreatedAt.Format("2006-01-02T15:04:05Z07:00"))
	fmt.Printf("  Tool: %s\n", m.Bundle.CreatedBy)

	// Check for signature
	sigPath := filepath.Join(path, "signature.sig")
	if _, err := os.Stat(sigPath); err == nil {
		if verify {
			fmt.Printf("  Signed: Yes (verification not yet implemented)\n")
		} else {
			fmt.Printf("  Signed: Yes\n")
		}
	} else {
		fmt.Printf("  Signed: No\n")
	}

	fmt.Println()
	fmt.Printf("Copyright: %d %s\n", m.Rights.CopyrightYear, m.Rights.CopyrightHolder)

	if showTracks {
		fmt.Println()
		fmt.Println("Track Listing:")
		fmt.Println("--------------")
		for _, track := range m.Tracks {
			duration := track.Duration
			if duration == "" {
				duration = "--:--"
			}
			fmt.Printf("  %2d. %s [%s]\n", track.Number, track.Title, duration)
		}
	}

	return nil
}

func calculateTotalDuration(tracks []manifest.Track) string {
	totalSeconds := 0

	for _, track := range tracks {
		if track.Duration == "" {
			continue
		}

		// Parse MM:SS format
		var min, sec int
		_, err := fmt.Sscanf(track.Duration, "%d:%d", &min, &sec)
		if err != nil {
			continue
		}
		totalSeconds += min*60 + sec
	}

	if totalSeconds == 0 {
		return ""
	}

	hours := totalSeconds / 3600
	minutes := (totalSeconds % 3600) / 60
	seconds := totalSeconds % 60

	if hours > 0 {
		return fmt.Sprintf("%d:%02d:%02d", hours, minutes, seconds)
	}
	return fmt.Sprintf("%d:%02d", minutes, seconds)
}
