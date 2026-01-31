package convert

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Converter handles WAV to MP3 conversion
type Converter struct {
	InputFiles []string
	OutputDir  string
	Bitrate    int  // CBR: 128, 192, 256, 320
	Quality    int  // VBR: 0-9, -1 means use CBR
	Force      bool
	Verbose    bool
}

// ConvertResult represents the result of a single file conversion
type ConvertResult struct {
	InputPath  string
	OutputPath string
	Success    bool
	Error      error
}

// NewConverter creates a new converter instance
func NewConverter(inputs []string, outputDir string, bitrate, quality int, force, verbose bool) *Converter {
	return &Converter{
		InputFiles: inputs,
		OutputDir:  outputDir,
		Bitrate:    bitrate,
		Quality:    quality,
		Force:      force,
		Verbose:    verbose,
	}
}

// Convert performs the conversion of all input files
func (c *Converter) Convert() ([]ConvertResult, error) {
	// Expand inputs to get all WAV files
	wavFiles, err := c.expandInputs()
	if err != nil {
		return nil, fmt.Errorf("failed to expand inputs: %w", err)
	}

	if len(wavFiles) == 0 {
		return nil, fmt.Errorf("no WAV files found")
	}

	// Initialize LAME runner
	lame, err := NewLameRunner()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize LAME: %w", err)
	}
	defer lame.Cleanup()

	// Determine mode string for output
	modeStr := c.getModeString()

	fmt.Println("Converting WAV files to MP3...")
	fmt.Println()

	results := make([]ConvertResult, 0, len(wavFiles))
	successCount := 0
	failCount := 0

	for i, wavFile := range wavFiles {
		result := c.convertFile(lame, wavFile, i+1, len(wavFiles), modeStr)
		results = append(results, result)
		if result.Success {
			successCount++
		} else {
			failCount++
		}
	}

	fmt.Println()
	if failCount == 0 {
		fmt.Printf("Converted %d file(s) successfully.\n", successCount)
	} else {
		fmt.Printf("Converted %d of %d file(s). %d failed.\n", successCount, len(wavFiles), failCount)
	}

	return results, nil
}

// expandInputs expands input paths to a list of WAV files
func (c *Converter) expandInputs() ([]string, error) {
	var wavFiles []string

	for _, input := range c.InputFiles {
		info, err := os.Stat(input)
		if err != nil {
			return nil, fmt.Errorf("cannot access %s: %w", input, err)
		}

		if info.IsDir() {
			// Find all WAV files in directory
			entries, err := os.ReadDir(input)
			if err != nil {
				return nil, fmt.Errorf("cannot read directory %s: %w", input, err)
			}
			for _, entry := range entries {
				if !entry.IsDir() && isWavFile(entry.Name()) {
					wavFiles = append(wavFiles, filepath.Join(input, entry.Name()))
				}
			}
		} else {
			// Single file
			if !isWavFile(input) {
				return nil, fmt.Errorf("not a WAV file: %s", input)
			}
			wavFiles = append(wavFiles, input)
		}
	}

	return wavFiles, nil
}

// convertFile converts a single WAV file to MP3
func (c *Converter) convertFile(lame *LameRunner, inputPath string, index, total int, modeStr string) ConvertResult {
	outputPath := c.getOutputPath(inputPath)
	inputName := filepath.Base(inputPath)
	outputName := filepath.Base(outputPath)

	// Print progress
	fmt.Printf("[%d/%d] %s → %s (%s)... ", index, total, inputName, outputName, modeStr)

	// Check if output exists
	if _, err := os.Stat(outputPath); err == nil {
		if !c.Force {
			fmt.Println("SKIPPED (exists)")
			return ConvertResult{
				InputPath:  inputPath,
				OutputPath: outputPath,
				Success:    false,
				Error:      fmt.Errorf("output file already exists (use --force to overwrite)"),
			}
		}
	}

	// Create output directory if needed
	outputDir := filepath.Dir(outputPath)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		fmt.Println("FAILED")
		fmt.Printf("  Error: %v\n", err)
		return ConvertResult{
			InputPath:  inputPath,
			OutputPath: outputPath,
			Success:    false,
			Error:      err,
		}
	}

	// Build LAME options
	opts := LameOptions{
		Bitrate: c.Bitrate,
		Quality: c.Quality,
	}

	// Run conversion
	if err := lame.Convert(inputPath, outputPath, opts); err != nil {
		fmt.Println("FAILED")
		fmt.Printf("  Error: %v\n", err)
		return ConvertResult{
			InputPath:  inputPath,
			OutputPath: outputPath,
			Success:    false,
			Error:      err,
		}
	}

	fmt.Println("done")
	return ConvertResult{
		InputPath:  inputPath,
		OutputPath: outputPath,
		Success:    true,
		Error:      nil,
	}
}

// getOutputPath determines the output path for a given input file
func (c *Converter) getOutputPath(inputPath string) string {
	baseName := strings.TrimSuffix(filepath.Base(inputPath), filepath.Ext(inputPath))
	outputName := baseName + ".mp3"

	if c.OutputDir != "" {
		return filepath.Join(c.OutputDir, outputName)
	}
	return filepath.Join(filepath.Dir(inputPath), outputName)
}

// getModeString returns a string describing the encoding mode
func (c *Converter) getModeString() string {
	if c.Quality >= 0 {
		return fmt.Sprintf("VBR V%d", c.Quality)
	}
	return fmt.Sprintf("%d kbps", c.Bitrate)
}

// isWavFile checks if a filename has a WAV extension
func isWavFile(name string) bool {
	ext := strings.ToLower(filepath.Ext(name))
	return ext == ".wav"
}

// ValidateBitrate checks if the bitrate is a valid value
func ValidateBitrate(bitrate int) error {
	validBitrates := []int{128, 192, 256, 320}
	for _, valid := range validBitrates {
		if bitrate == valid {
			return nil
		}
	}
	return fmt.Errorf("invalid bitrate %d: must be one of 128, 192, 256, 320", bitrate)
}

// ValidateQuality checks if the VBR quality is a valid value
func ValidateQuality(quality int) error {
	if quality < 0 || quality > 9 {
		return fmt.Errorf("invalid quality %d: must be 0-9 (lower is better)", quality)
	}
	return nil
}
