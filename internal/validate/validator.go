package validate

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/davesmith10/rice-cli/pkg/manifest"
	"gopkg.in/yaml.v3"
)

// Result represents a validation result
type Result struct {
	Category string
	Check    string
	Passed   bool
	Severity string // "error" or "warning"
	Message  string
}

// Report contains all validation results
type Report struct {
	Path    string
	Results []Result
	Errors  int
	Warns   int
}

// Validator validates ricecake bundles
type Validator struct {
	path     string
	strict   bool
	manifest *manifest.Manifest
	report   *Report
}

// New creates a new validator
func New(path string, strict bool) *Validator {
	return &Validator{
		path:   path,
		strict: strict,
		report: &Report{Path: path},
	}
}

// Validate runs all validation checks
func (v *Validator) Validate() (*Report, error) {
	// Check if path exists
	info, err := os.Stat(v.path)
	if err != nil {
		return nil, fmt.Errorf("path does not exist: %s", v.path)
	}

	// Handle both directories and .ricecake files
	if !info.IsDir() {
		if strings.HasSuffix(v.path, ".ricecake") {
			return nil, fmt.Errorf("validation of .ricecake files not yet implemented - validate the source directory instead")
		}
		return nil, fmt.Errorf("path must be a directory: %s", v.path)
	}

	// Run validation checks
	v.validateStructure()
	v.validateManifest()
	v.validateAudio()
	v.validateImages()
	v.validateSecurity()
	v.validateCopyright()

	return v.report, nil
}

func (v *Validator) addResult(category, check string, passed bool, severity, message string) {
	result := Result{
		Category: category,
		Check:    check,
		Passed:   passed,
		Severity: severity,
		Message:  message,
	}
	v.report.Results = append(v.report.Results, result)

	if !passed {
		if severity == "error" {
			v.report.Errors++
		} else {
			v.report.Warns++
		}
	}
}

func (v *Validator) validateStructure() {
	// Check manifest.yaml exists
	manifestPath := filepath.Join(v.path, "manifest.yaml")
	if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
		v.addResult("Structure", "manifest.yaml exists", false, "error", "manifest.yaml not found")
	} else {
		v.addResult("Structure", "manifest.yaml exists", true, "", "")
	}

	// Check copyright.txt exists
	copyrightPath := filepath.Join(v.path, "copyright.txt")
	if _, err := os.Stat(copyrightPath); os.IsNotExist(err) {
		v.addResult("Structure", "copyright.txt exists", false, "error", "copyright.txt not found")
	} else {
		v.addResult("Structure", "copyright.txt exists", true, "", "")
	}

	// Check audio/ directory exists
	audioDir := filepath.Join(v.path, "audio")
	if info, err := os.Stat(audioDir); os.IsNotExist(err) || !info.IsDir() {
		v.addResult("Structure", "audio/ directory exists", false, "error", "audio/ directory not found")
	} else {
		v.addResult("Structure", "audio/ directory exists", true, "", "")
	}

	// Check images/ directory exists
	imagesDir := filepath.Join(v.path, "images")
	if info, err := os.Stat(imagesDir); os.IsNotExist(err) || !info.IsDir() {
		v.addResult("Structure", "images/ directory exists", false, "error", "images/ directory not found")
	} else {
		v.addResult("Structure", "images/ directory exists", true, "", "")
	}

	// Check cover image exists
	coverFound := false
	for _, ext := range []string{".jpg", ".jpeg"} {
		coverPath := filepath.Join(v.path, "images", "cover"+ext)
		if _, err := os.Stat(coverPath); err == nil {
			coverFound = true
			break
		}
	}
	if !coverFound {
		v.addResult("Structure", "cover image exists", false, "error", "cover.jpg not found in images/")
	} else {
		v.addResult("Structure", "cover image exists", true, "", "")
	}
}

func (v *Validator) validateManifest() {
	manifestPath := filepath.Join(v.path, "manifest.yaml")
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		v.addResult("Manifest", "readable", false, "error", fmt.Sprintf("cannot read manifest: %v", err))
		return
	}

	var m manifest.Manifest
	if err := yaml.Unmarshal(data, &m); err != nil {
		v.addResult("Manifest", "valid YAML syntax", false, "error", fmt.Sprintf("invalid YAML: %v", err))
		return
	}
	v.addResult("Manifest", "valid YAML syntax", true, "", "")
	v.manifest = &m

	// Check required fields
	if m.ManifestVersion == 0 {
		v.addResult("Manifest", "manifest_version present", false, "error", "manifest_version is required")
	} else {
		v.addResult("Manifest", "manifest_version present", true, "", "")
	}

	if m.Release.Title == "" {
		v.addResult("Manifest", "release.title present", false, "error", "release.title is required")
	} else {
		v.addResult("Manifest", "release.title present", true, "", "")
	}

	if m.Release.Artist == "" {
		v.addResult("Manifest", "release.artist present", false, "error", "release.artist is required")
	} else {
		v.addResult("Manifest", "release.artist present", true, "", "")
	}

	if m.Release.ReleaseDate == "" {
		v.addResult("Manifest", "release.release_date present", false, "error", "release.release_date is required")
	} else {
		v.addResult("Manifest", "release.release_date present", true, "", "")
	}

	if len(m.Tracks) == 0 {
		v.addResult("Manifest", "tracks present", false, "error", "at least one track is required")
	} else {
		v.addResult("Manifest", "tracks present", true, "", "")

		// Validate each track
		for i, track := range m.Tracks {
			if track.Number == 0 {
				v.addResult("Manifest", fmt.Sprintf("track[%d].number", i), false, "error", "track number is required")
			}
			if track.Title == "" {
				v.addResult("Manifest", fmt.Sprintf("track[%d].title", i), false, "error", "track title is required")
			}
			if track.Filename == "" {
				v.addResult("Manifest", fmt.Sprintf("track[%d].filename", i), false, "error", "track filename is required")
			}
		}
	}

	if len(m.AudioFormats) == 0 {
		v.addResult("Manifest", "audio_formats present", false, "error", "at least one audio format is required")
	} else {
		v.addResult("Manifest", "audio_formats present", true, "", "")
	}

	if m.Images.Cover.Filename == "" {
		v.addResult("Manifest", "images.cover present", false, "error", "images.cover is required")
	} else {
		v.addResult("Manifest", "images.cover present", true, "", "")
	}

	if m.Rights.CopyrightYear == 0 {
		v.addResult("Manifest", "rights.copyright_year present", false, "error", "rights.copyright_year is required")
	} else {
		v.addResult("Manifest", "rights.copyright_year present", true, "", "")
	}

	if m.Rights.CopyrightHolder == "" {
		v.addResult("Manifest", "rights.copyright_holder present", false, "error", "rights.copyright_holder is required")
	} else {
		v.addResult("Manifest", "rights.copyright_holder present", true, "", "")
	}

	if m.Bundle.BundleID == "" {
		v.addResult("Manifest", "bundle.bundle_id present", false, "error", "bundle.bundle_id is required")
	} else {
		v.addResult("Manifest", "bundle.bundle_id present", true, "", "")
	}
}

func (v *Validator) validateAudio() {
	audioDir := filepath.Join(v.path, "audio")
	if _, err := os.Stat(audioDir); os.IsNotExist(err) {
		return // Already reported in structure check
	}

	entries, err := os.ReadDir(audioDir)
	if err != nil {
		v.addResult("Audio", "readable", false, "error", fmt.Sprintf("cannot read audio directory: %v", err))
		return
	}

	audioCount := 0
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		ext := strings.ToLower(filepath.Ext(name))

		// Check if extension is allowed
		if !manifest.AllowedAudioExtensions[ext] {
			v.addResult("Audio", fmt.Sprintf("file %s", name), false, "error", fmt.Sprintf("file type %s not allowed", ext))
			continue
		}

		// Check file size
		info, err := entry.Info()
		if err != nil {
			v.addResult("Audio", fmt.Sprintf("file %s", name), false, "error", fmt.Sprintf("cannot get file info: %v", err))
			continue
		}

		if info.Size() > manifest.MaxSingleAudioFile {
			v.addResult("Audio", fmt.Sprintf("file %s size", name), false, "error",
				fmt.Sprintf("file exceeds maximum size of %d MB", manifest.MaxSingleAudioFile/(1024*1024)))
			continue
		}

		// Verify magic bytes
		if err := v.verifyMagicBytes(filepath.Join(audioDir, name), ext); err != nil {
			v.addResult("Audio", fmt.Sprintf("file %s magic bytes", name), false, "error", err.Error())
			continue
		}

		audioCount++
		v.addResult("Audio", fmt.Sprintf("file %s", name), true, "", "")
	}

	if audioCount == 0 {
		v.addResult("Audio", "at least one audio file", false, "error", "no valid audio files found")
	}
}

func (v *Validator) validateImages() {
	imagesDir := filepath.Join(v.path, "images")
	if _, err := os.Stat(imagesDir); os.IsNotExist(err) {
		return // Already reported in structure check
	}

	entries, err := os.ReadDir(imagesDir)
	if err != nil {
		v.addResult("Images", "readable", false, "error", fmt.Sprintf("cannot read images directory: %v", err))
		return
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		ext := strings.ToLower(filepath.Ext(name))

		// Skip README files
		if ext == ".txt" {
			continue
		}

		// Check if extension is allowed
		if !manifest.AllowedImageExtensions[ext] {
			v.addResult("Images", fmt.Sprintf("file %s", name), false, "error", fmt.Sprintf("file type %s not allowed", ext))
			continue
		}

		// Check file size
		info, err := entry.Info()
		if err != nil {
			v.addResult("Images", fmt.Sprintf("file %s", name), false, "error", fmt.Sprintf("cannot get file info: %v", err))
			continue
		}

		if info.Size() > manifest.MaxSingleImageFile {
			v.addResult("Images", fmt.Sprintf("file %s size", name), false, "error",
				fmt.Sprintf("file exceeds maximum size of %d MB", manifest.MaxSingleImageFile/(1024*1024)))
			continue
		}

		// Verify magic bytes
		if err := v.verifyMagicBytes(filepath.Join(imagesDir, name), ext); err != nil {
			v.addResult("Images", fmt.Sprintf("file %s magic bytes", name), false, "error", err.Error())
			continue
		}

		// Check cover image dimensions
		if strings.HasPrefix(strings.ToLower(name), "cover") {
			if err := v.validateCoverDimensions(filepath.Join(imagesDir, name)); err != nil {
				v.addResult("Images", fmt.Sprintf("file %s dimensions", name), false, "error", err.Error())
				continue
			}
		}

		v.addResult("Images", fmt.Sprintf("file %s", name), true, "", "")
	}
}

func (v *Validator) validateSecurity() {
	err := filepath.Walk(v.path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip the root directory
		if path == v.path {
			return nil
		}

		relPath, _ := filepath.Rel(v.path, path)

		// Check for path traversal
		if strings.Contains(relPath, "..") {
			v.addResult("Security", fmt.Sprintf("path %s", relPath), false, "error", "path traversal detected")
			return nil
		}

		// Check for hidden files (but allow directories)
		if !info.IsDir() && strings.HasPrefix(filepath.Base(path), ".") {
			v.addResult("Security", fmt.Sprintf("file %s", relPath), false, "error", "hidden files not allowed")
			return nil
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Check that all files have extensions
		ext := filepath.Ext(path)
		if ext == "" {
			v.addResult("Security", fmt.Sprintf("file %s", relPath), false, "error", "files must have extensions")
			return nil
		}

		// Check against whitelist
		ext = strings.ToLower(ext)
		if !manifest.AllAllowedExtensions[ext] {
			v.addResult("Security", fmt.Sprintf("file %s", relPath), false, "error", fmt.Sprintf("file type %s not allowed", ext))
			return nil
		}

		return nil
	})

	if err != nil {
		v.addResult("Security", "file scan", false, "error", fmt.Sprintf("error scanning files: %v", err))
	}
}

func (v *Validator) validateCopyright() {
	copyrightPath := filepath.Join(v.path, "copyright.txt")
	data, err := os.ReadFile(copyrightPath)
	if err != nil {
		return // Already reported in structure check
	}

	content := string(data)

	if len(content) == 0 {
		v.addResult("Copyright", "file not empty", false, "error", "copyright.txt is empty")
		return
	}
	v.addResult("Copyright", "file not empty", true, "", "")

	if !strings.Contains(strings.ToLower(content), "copyright") {
		v.addResult("Copyright", "contains copyright declaration", false, "error", "must contain 'Copyright' declaration")
	} else {
		v.addResult("Copyright", "contains copyright declaration", true, "", "")
	}

	// Check for copyright holder (look for common patterns)
	hasHolder := strings.Contains(content, "Copyright Holder:") ||
		strings.Contains(content, "copyright holder:") ||
		strings.Contains(content, "©")
	if !hasHolder {
		v.addResult("Copyright", "contains copyright holder", false, "error", "must specify copyright holder")
	} else {
		v.addResult("Copyright", "contains copyright holder", true, "", "")
	}

	// Check file size
	if len(data) > 10*1024 {
		v.addResult("Copyright", "file size", false, "error", "copyright.txt exceeds 10 KB limit")
	}
}

func (v *Validator) verifyMagicBytes(path, ext string) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("cannot open file: %v", err)
	}
	defer file.Close()

	// Read first 12 bytes for magic detection
	header := make([]byte, 12)
	n, err := file.Read(header)
	if err != nil || n < 4 {
		return fmt.Errorf("cannot read file header")
	}

	switch ext {
	case ".mp3":
		// Check for ID3 tag or MP3 frame sync
		if !(header[0] == 0x49 && header[1] == 0x44 && header[2] == 0x33) && // ID3
			!(header[0] == 0xFF && (header[1]&0xE0) == 0xE0) { // Frame sync
			return fmt.Errorf("file does not appear to be a valid MP3")
		}
	case ".flac":
		// Check for fLaC magic
		if !(header[0] == 0x66 && header[1] == 0x4C && header[2] == 0x61 && header[3] == 0x43) {
			return fmt.Errorf("file does not appear to be a valid FLAC")
		}
	case ".jpg", ".jpeg":
		// Check for JPEG magic
		if !(header[0] == 0xFF && header[1] == 0xD8 && header[2] == 0xFF) {
			return fmt.Errorf("file does not appear to be a valid JPEG")
		}
	case ".ogg":
		// Check for OggS magic
		if !(header[0] == 0x4F && header[1] == 0x67 && header[2] == 0x67 && header[3] == 0x53) {
			return fmt.Errorf("file does not appear to be a valid OGG")
		}
	case ".wav":
		// Check for RIFF magic
		if !(header[0] == 0x52 && header[1] == 0x49 && header[2] == 0x46 && header[3] == 0x46) {
			return fmt.Errorf("file does not appear to be a valid WAV")
		}
	}

	return nil
}

func (v *Validator) validateCoverDimensions(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("cannot open file: %v", err)
	}
	defer file.Close()

	// We'll use a simple JPEG dimension check
	// Skip to SOF0 marker to get dimensions
	// This is a simplified check - for production we'd use image.DecodeConfig

	// For now, just check file exists and has reasonable size
	info, err := os.Stat(path)
	if err != nil {
		return err
	}

	// A 1400x1400 JPEG should be at least ~50KB typically
	if info.Size() < 10*1024 {
		return fmt.Errorf("cover image appears too small (file size suggests low resolution)")
	}

	return nil
}
