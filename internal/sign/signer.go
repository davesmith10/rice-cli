package sign

import (
	"archive/zip"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// Signer handles bundle signing operations
type Signer struct {
	privateKey ed25519.PrivateKey
	verbose    bool
}

// NewSigner creates a new signer with the given private key
func NewSigner(privateKey ed25519.PrivateKey, verbose bool) *Signer {
	return &Signer{
		privateKey: privateKey,
		verbose:    verbose,
	}
}

// LoadPrivateKey loads an Ed25519 private key from a file
func LoadPrivateKey(path string) (ed25519.PrivateKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read key file: %w", err)
	}

	block, _ := pem.Decode(data)
	if block == nil {
		// Try raw base64
		decoded, err := base64.StdEncoding.DecodeString(strings.TrimSpace(string(data)))
		if err != nil {
			return nil, fmt.Errorf("failed to decode key: not PEM or base64")
		}
		if len(decoded) != ed25519.PrivateKeySize {
			return nil, fmt.Errorf("invalid key size: expected %d bytes, got %d", ed25519.PrivateKeySize, len(decoded))
		}
		return ed25519.PrivateKey(decoded), nil
	}

	if block.Type != "PRIVATE KEY" && block.Type != "ED25519 PRIVATE KEY" {
		return nil, fmt.Errorf("unexpected key type: %s", block.Type)
	}

	if len(block.Bytes) != ed25519.PrivateKeySize {
		return nil, fmt.Errorf("invalid key size: expected %d bytes, got %d", ed25519.PrivateKeySize, len(block.Bytes))
	}

	return ed25519.PrivateKey(block.Bytes), nil
}

// LoadPrivateKeyFromEnv loads a private key from an environment variable
func LoadPrivateKeyFromEnv(envVar string) (ed25519.PrivateKey, error) {
	value := os.Getenv(envVar)
	if value == "" {
		return nil, fmt.Errorf("environment variable %s is not set", envVar)
	}

	decoded, err := base64.StdEncoding.DecodeString(value)
	if err != nil {
		return nil, fmt.Errorf("failed to decode key from environment: %w", err)
	}

	if len(decoded) != ed25519.PrivateKeySize {
		return nil, fmt.Errorf("invalid key size: expected %d bytes, got %d", ed25519.PrivateKeySize, len(decoded))
	}

	return ed25519.PrivateKey(decoded), nil
}

// SignBundle signs a ricecake bundle
func (s *Signer) SignBundle(bundlePath string) error {
	// Create temp directory for extraction
	tempDir, err := os.MkdirTemp("", "rice-sign-*")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	// Extract bundle
	if s.verbose {
		fmt.Println("Extracting bundle...")
	}
	if err := extractBundle(bundlePath, tempDir); err != nil {
		return fmt.Errorf("failed to extract bundle: %w", err)
	}

	// Read manifest to get bundle ID
	manifestPath := filepath.Join(tempDir, "manifest.yaml")
	manifestData, err := os.ReadFile(manifestPath)
	if err != nil {
		return fmt.Errorf("failed to read manifest: %w", err)
	}

	var manifest struct {
		Bundle struct {
			BundleID string `yaml:"bundle_id"`
		} `yaml:"bundle"`
	}
	if err := yaml.Unmarshal(manifestData, &manifest); err != nil {
		return fmt.Errorf("failed to parse manifest: %w", err)
	}

	// Compute content hash
	if s.verbose {
		fmt.Println("Computing content hash...")
	}
	contentHash, err := s.computeContentHash(tempDir)
	if err != nil {
		return fmt.Errorf("failed to compute content hash: %w", err)
	}

	// Generate signature
	if s.verbose {
		fmt.Println("Generating signature...")
	}
	signature := ed25519.Sign(s.privateKey, contentHash)

	// Create signature file
	sigContent := s.createSignatureFile(manifest.Bundle.BundleID, contentHash, signature)
	sigPath := filepath.Join(tempDir, "signature.sig")
	if err := os.WriteFile(sigPath, []byte(sigContent), 0644); err != nil {
		return fmt.Errorf("failed to write signature file: %w", err)
	}

	// Rebuild bundle
	if s.verbose {
		fmt.Println("Rebuilding bundle with signature...")
	}

	// Create new bundle path
	newBundlePath := bundlePath + ".tmp"
	if err := rebuildBundle(tempDir, newBundlePath); err != nil {
		return fmt.Errorf("failed to rebuild bundle: %w", err)
	}

	// Replace original bundle
	if err := os.Remove(bundlePath); err != nil {
		os.Remove(newBundlePath)
		return fmt.Errorf("failed to remove original bundle: %w", err)
	}
	if err := os.Rename(newBundlePath, bundlePath); err != nil {
		return fmt.Errorf("failed to rename bundle: %w", err)
	}

	return nil
}

func (s *Signer) computeContentHash(dir string) ([]byte, error) {
	hash := sha256.New()

	// Get all files sorted for deterministic hashing
	var files []string
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		// Skip signature file if it exists
		if filepath.Base(path) == "signature.sig" {
			return nil
		}
		relPath, _ := filepath.Rel(dir, path)
		files = append(files, relPath)
		return nil
	})
	if err != nil {
		return nil, err
	}

	sort.Strings(files)

	for _, relPath := range files {
		fullPath := filepath.Join(dir, relPath)

		// Hash the relative path
		hash.Write([]byte(relPath))

		// Hash the file contents
		file, err := os.Open(fullPath)
		if err != nil {
			return nil, err
		}
		if _, err := io.Copy(hash, file); err != nil {
			file.Close()
			return nil, err
		}
		file.Close()

		if s.verbose {
			fmt.Printf("  Hashing %s\n", relPath)
		}
	}

	return hash.Sum(nil), nil
}

func (s *Signer) createSignatureFile(bundleID string, contentHash, signature []byte) string {
	return fmt.Sprintf(`-----BEGIN RICECAKE SIGNATURE-----
Version: 1
Bundle-ID: %s
Created-At: %s
Tool-Version: rice-cli v1.0.0
Hash-Algorithm: SHA-256
Content-Hash: %s

%s
-----END RICECAKE SIGNATURE-----
`,
		bundleID,
		time.Now().UTC().Format(time.RFC3339),
		base64.StdEncoding.EncodeToString(contentHash),
		base64.StdEncoding.EncodeToString(signature),
	)
}

func extractBundle(bundlePath, destDir string) error {
	reader, err := zip.OpenReader(bundlePath)
	if err != nil {
		return err
	}
	defer reader.Close()

	for _, file := range reader.File {
		destPath := filepath.Join(destDir, file.Name)

		// Security check
		if !strings.HasPrefix(destPath, filepath.Clean(destDir)+string(os.PathSeparator)) {
			return fmt.Errorf("invalid file path: %s", file.Name)
		}

		if file.FileInfo().IsDir() {
			os.MkdirAll(destPath, 0755)
			continue
		}

		// Create parent directories
		os.MkdirAll(filepath.Dir(destPath), 0755)

		// Extract file
		srcFile, err := file.Open()
		if err != nil {
			return err
		}

		destFile, err := os.Create(destPath)
		if err != nil {
			srcFile.Close()
			return err
		}

		_, err = io.Copy(destFile, srcFile)
		srcFile.Close()
		destFile.Close()
		if err != nil {
			return err
		}
	}

	return nil
}

func rebuildBundle(sourceDir, destPath string) error {
	outFile, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer outFile.Close()

	zipWriter := zip.NewWriter(outFile)
	defer zipWriter.Close()

	return filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return err
		}

		if relPath == "." {
			return nil
		}

		// Convert to forward slashes
		zipPath := strings.ReplaceAll(relPath, string(filepath.Separator), "/")

		if info.IsDir() {
			_, err = zipWriter.Create(zipPath + "/")
			return err
		}

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

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		_, err = io.Copy(writer, file)
		return err
	})
}

// GenerateKeyPair generates a new Ed25519 key pair
func GenerateKeyPair() (ed25519.PublicKey, ed25519.PrivateKey, error) {
	return ed25519.GenerateKey(rand.Reader)
}

// SaveKeyPair saves a key pair to files
func SaveKeyPair(publicKey ed25519.PublicKey, privateKey ed25519.PrivateKey, outputDir string) error {
	// Save private key
	privateBlock := &pem.Block{
		Type:  "ED25519 PRIVATE KEY",
		Bytes: privateKey,
	}
	privatePath := filepath.Join(outputDir, "private.key")
	if err := os.WriteFile(privatePath, pem.EncodeToMemory(privateBlock), 0600); err != nil {
		return fmt.Errorf("failed to save private key: %w", err)
	}

	// Save public key
	publicBlock := &pem.Block{
		Type:  "ED25519 PUBLIC KEY",
		Bytes: publicKey,
	}
	publicPath := filepath.Join(outputDir, "public.key")
	if err := os.WriteFile(publicPath, pem.EncodeToMemory(publicBlock), 0644); err != nil {
		return fmt.Errorf("failed to save public key: %w", err)
	}

	return nil
}
