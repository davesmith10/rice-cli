# Rice CLI

A command-line tool for creating, validating, and managing ricecake bundles.

A **ricecake** is a self-contained music release bundle distributed as a single archive file (`.ricecake`). It packages audio files, artwork, metadata, and liner notes into a validated, optionally signed bundle.

## Installation

### From Source

```bash
go install github.com/davesmith10/rice-cli/cmd/rice@latest
```

### Build Locally

```bash
git clone https://github.com/davesmith10/rice-cli.git
cd rice-cli
go build -o rice ./cmd/rice/
```

## Quick Start

```bash
# Initialize a new bundle directory
rice init my-album --artist "Artist Name" --title "Album Title" --tracks 10

# Add your audio files to my-album/audio/
# Add cover.jpg to my-album/images/
# Edit my-album/manifest.yaml with track details

# Validate the bundle
rice validate my-album/

# Build the bundle
rice build my-album/

# Preview locally
rice test my-album/ --open
```

## Commands

### `rice init`

Initialize a new bundle directory with template files.

```bash
rice init [directory] [flags]

Flags:
  --artist string    Artist or band name
  --title string     Album or release title
  --tracks int       Number of tracks to template (default 1)
```

### `rice build`

Create a ricecake bundle from a directory.

```bash
rice build [directory] [flags]

Flags:
  --output string    Output bundle path (default: [directory].ricecake)
  --no-validate      Skip validation before building
  --force            Overwrite existing bundle
```

### `rice validate`

Validate a bundle or directory against the ricecake specification.

```bash
rice validate [path] [flags]

Flags:
  --strict    Enable strict validation (warnings become errors)
  --json      Output results as JSON
  --quiet     Only output errors
```

### `rice sign`

Add a digital signature to a bundle using Ed25519.

```bash
rice sign [bundle] [flags]

Flags:
  --key string       Path to private key file
  --key-env string   Environment variable containing key (default: RICE_SIGNING_KEY)
```

### `rice test`

Start a local preview server to test a bundle.

```bash
rice test [bundle-or-directory] [flags]

Flags:
  --port int    Server port (default 8080)
  --open        Open browser automatically
```

### `rice info`

Display information about a bundle.

```bash
rice info [bundle] [flags]

Flags:
  --json      Output as JSON
  --tracks    Show detailed track listing
  --verify    Verify signature if present
```

### `rice describe`

Print the raw manifest.yaml contents of a bundle.

```bash
rice describe [bundle-or-directory] [flags]

Flags:
  --raw    Output raw YAML without header comment
```

### `rice keygen`

Generate an Ed25519 key pair for signing bundles.

```bash
rice keygen [flags]

Flags:
  --output string    Output directory for keys (default: ~/.rice/)
```

### `rice convert`

Convert WAV files to high-quality MP3. Useful for preparing audio files before adding them to a bundle.

```bash
rice convert [files...] [flags]

Flags:
  -o, --output string   Output directory (default: same as input)
  -b, --bitrate int     CBR bitrate: 128, 192, 256, 320 kbps (default: 320)
  -q, --quality int     VBR quality: 0-9, lower is better (overrides --bitrate)
  -f, --force           Overwrite existing output files
```

**Examples:**

```bash
# Convert a single file at maximum quality (320 kbps)
rice convert recording.wav

# Convert multiple files
rice convert track1.wav track2.wav track3.wav

# Convert all WAV files in a directory
rice convert audio/

# Use variable bitrate for smaller file sizes
rice convert *.wav --quality 2

# Output to a specific directory
rice convert *.wav --output converted/

# Use 256 kbps bitrate
rice convert track.wav --bitrate 256
```

**Notes:**
- The LAME encoder is embedded in the binary - no external dependencies required
- Default mode is CBR (constant bitrate) at 320 kbps for maximum quality
- VBR (variable bitrate) mode with `--quality 2` produces high-quality files with smaller sizes
- Currently supports Linux only; Windows support planned for future release

## Bundle Structure

```
my-album.ricecake/
├── manifest.yaml           # Bundle metadata (required)
├── copyright.txt           # Copyright declaration (required)
├── signature.sig           # Digital signature (optional)
├── audio/                  # Audio files (required)
│   ├── 001-track-name.mp3
│   ├── 001-track-name.flac
│   └── ...
├── images/                 # Visual assets (required)
│   ├── cover.jpg           # Album cover (required, min 1400x1400)
│   └── ...
└── liner-notes/            # Written content (optional)
    ├── notes.txt
    └── credits.txt
```

## Supported Formats

### Audio
- MP3 (128-320 kbps)
- FLAC (16/24-bit, 44.1-96 kHz)
- OGG (Vorbis)
- WAV

### Images
- JPEG only (for cover art)

## License

All Rights Reserved.
