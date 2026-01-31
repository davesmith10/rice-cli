package manifest

import "time"

// Manifest represents the complete manifest.yaml structure
type Manifest struct {
	ManifestVersion int           `yaml:"manifest_version"`
	Release         Release       `yaml:"release"`
	Tracks          []Track       `yaml:"tracks"`
	AudioFormats    []AudioFormat `yaml:"audio_formats"`
	Images          Images        `yaml:"images"`
	Rights          Rights        `yaml:"rights"`
	Bundle          BundleInfo    `yaml:"bundle"`
}

// Release contains album/release information
type Release struct {
	Title         string `yaml:"title"`
	Artist        string `yaml:"artist"`
	ReleaseDate   string `yaml:"release_date"`
	Genre         string `yaml:"genre"`
	Subgenre      string `yaml:"subgenre,omitempty"`
	CatalogNumber string `yaml:"catalog_number,omitempty"`
}

// Track represents a single track in the release
type Track struct {
	Number     int      `yaml:"number"`
	Title      string   `yaml:"title"`
	Duration   string   `yaml:"duration,omitempty"`
	Filename   string   `yaml:"filename"`
	Composers  []string `yaml:"composers,omitempty"`
	Performers []string `yaml:"performers,omitempty"`
}

// AudioFormat describes an available audio format
type AudioFormat struct {
	Format     string `yaml:"format"`
	Bitrate    int    `yaml:"bitrate,omitempty"`
	BitDepth   int    `yaml:"bit_depth,omitempty"`
	SampleRate int    `yaml:"sample_rate,omitempty"`
}

// Images contains image asset information
type Images struct {
	Cover      ImageInfo  `yaml:"cover"`
	CoverLarge *ImageInfo `yaml:"cover_large,omitempty"`
	Back       *ImageInfo `yaml:"back,omitempty"`
	Artist     *ImageInfo `yaml:"artist,omitempty"`
}

// ImageInfo describes an image file
type ImageInfo struct {
	Filename string `yaml:"filename"`
	Width    int    `yaml:"width,omitempty"`
	Height   int    `yaml:"height,omitempty"`
}

// Rights contains copyright and licensing information
type Rights struct {
	CopyrightYear   int    `yaml:"copyright_year"`
	CopyrightHolder string `yaml:"copyright_holder"`
	License         string `yaml:"license,omitempty"`
	Contact         string `yaml:"contact,omitempty"`
}

// BundleInfo contains bundle metadata
type BundleInfo struct {
	CreatedBy string    `yaml:"created_by"`
	CreatedAt time.Time `yaml:"created_at"`
	BundleID  string    `yaml:"bundle_id"`
}

// AllowedAudioExtensions lists permitted audio file extensions
var AllowedAudioExtensions = map[string]bool{
	".mp3":  true,
	".flac": true,
	".ogg":  true,
	".wav":  true,
}

// AllowedImageExtensions lists permitted image file extensions
var AllowedImageExtensions = map[string]bool{
	".jpg":  true,
	".jpeg": true,
}

// AllowedTextExtensions lists permitted text file extensions
var AllowedTextExtensions = map[string]bool{
	".txt":  true,
	".yaml": true,
	".yml":  true,
}

// AllowedSignatureExtensions lists permitted signature file extensions
var AllowedSignatureExtensions = map[string]bool{
	".sig": true,
}

// AllAllowedExtensions combines all allowed extensions
var AllAllowedExtensions = func() map[string]bool {
	m := make(map[string]bool)
	for k, v := range AllowedAudioExtensions {
		m[k] = v
	}
	for k, v := range AllowedImageExtensions {
		m[k] = v
	}
	for k, v := range AllowedTextExtensions {
		m[k] = v
	}
	for k, v := range AllowedSignatureExtensions {
		m[k] = v
	}
	return m
}()

// Size limits
const (
	MaxSingleAudioFile  = 200 * 1024 * 1024 // 200 MB
	MaxSingleImageFile  = 20 * 1024 * 1024  // 20 MB
	MaxSingleTextFile   = 100 * 1024        // 100 KB
	MaxTotalBundleSize  = 2 * 1024 * 1024 * 1024 // 2 GB
	MaxTracks           = 99
	MaxFiles            = 500
	MinCoverDimension   = 1400
)
