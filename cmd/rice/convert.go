package main

import (
	"fmt"

	"github.com/davesmith10/rice-cli/internal/convert"
	"github.com/spf13/cobra"
)

func convertCmd() *cobra.Command {
	var outputDir string
	var bitrate int
	var quality int
	var force bool

	cmd := &cobra.Command{
		Use:   "convert [files...]",
		Short: "Convert WAV files to MP3",
		Long: `Convert one or more WAV files to high-quality MP3 format.

Supports both constant bitrate (CBR) and variable bitrate (VBR) modes.
By default, uses CBR at 320 kbps for maximum quality.

Examples:
  rice convert track.wav                     # Single file
  rice convert *.wav                         # Multiple files (shell expansion)
  rice convert audio/                        # All WAV files in directory
  rice convert track.wav --bitrate 256       # Lower bitrate
  rice convert *.wav --output converted/     # Output to specific directory
  rice convert *.wav --quality 2             # VBR mode (0-9, lower is better)`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runConvert(args, outputDir, bitrate, quality, force)
		},
	}

	cmd.Flags().StringVarP(&outputDir, "output", "o", "", "Output directory (default: same as input)")
	cmd.Flags().IntVarP(&bitrate, "bitrate", "b", 320, "CBR bitrate: 128, 192, 256, 320 kbps")
	cmd.Flags().IntVarP(&quality, "quality", "q", -1, "VBR quality: 0-9 (lower is better, overrides --bitrate)")
	cmd.Flags().BoolVarP(&force, "force", "f", false, "Overwrite existing output files")

	return cmd
}

func runConvert(inputs []string, outputDir string, bitrate, quality int, force bool) error {
	// Validate flags
	if quality >= 0 {
		if err := convert.ValidateQuality(quality); err != nil {
			return err
		}
	} else {
		if err := convert.ValidateBitrate(bitrate); err != nil {
			return err
		}
	}

	// Create converter
	converter := convert.NewConverter(inputs, outputDir, bitrate, quality, force, verbose)

	// Run conversion
	results, err := converter.Convert()
	if err != nil {
		return err
	}

	// Check for any failures
	for _, result := range results {
		if !result.Success {
			return fmt.Errorf("some conversions failed")
		}
	}

	return nil
}
