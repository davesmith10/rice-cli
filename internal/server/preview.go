package server

import (
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/davesmith10/rice-cli/pkg/manifest"
	"gopkg.in/yaml.v3"
)

// PreviewServer serves bundle content for testing
type PreviewServer struct {
	bundlePath string
	port       int
	verbose    bool
}

// NewPreviewServer creates a new preview server
func NewPreviewServer(bundlePath string, port int, verbose bool) *PreviewServer {
	return &PreviewServer{
		bundlePath: bundlePath,
		port:       port,
		verbose:    verbose,
	}
}

// Start starts the preview server
func (s *PreviewServer) Start() error {
	mux := http.NewServeMux()

	// Serve the preview page
	mux.HandleFunc("/", s.handleIndex)

	// Serve bundle files
	mux.HandleFunc("/files/", s.handleFiles)

	// Serve manifest as JSON
	mux.HandleFunc("/api/manifest", s.handleManifest)

	addr := fmt.Sprintf(":%d", s.port)
	fmt.Printf("Preview available at: http://localhost%s\n", addr)
	fmt.Println()
	fmt.Println("Press Ctrl+C to stop server.")
	fmt.Println()

	return http.ListenAndServe(addr, s.logMiddleware(mux))
}

func (s *PreviewServer) logMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if s.verbose {
			log.Printf("[INFO] %s %s", r.Method, r.URL.Path)
		}
		next.ServeHTTP(w, r)
	})
}

func (s *PreviewServer) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	// Load manifest
	manifestPath := filepath.Join(s.bundlePath, "manifest.yaml")
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		http.Error(w, "Failed to read manifest", http.StatusInternalServerError)
		return
	}

	var m manifest.Manifest
	if err := yaml.Unmarshal(data, &m); err != nil {
		http.Error(w, "Failed to parse manifest", http.StatusInternalServerError)
		return
	}

	// Render preview page
	tmpl := template.Must(template.New("preview").Parse(previewHTML))
	tmpl.Execute(w, m)
}

func (s *PreviewServer) handleFiles(w http.ResponseWriter, r *http.Request) {
	// Get requested file path
	filePath := strings.TrimPrefix(r.URL.Path, "/files/")
	if filePath == "" {
		http.NotFound(w, r)
		return
	}

	// Security: prevent path traversal
	if strings.Contains(filePath, "..") {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}

	fullPath := filepath.Join(s.bundlePath, filePath)

	// Check file exists
	info, err := os.Stat(fullPath)
	if err != nil || info.IsDir() {
		http.NotFound(w, r)
		return
	}

	// Set content type
	ext := strings.ToLower(filepath.Ext(fullPath))
	switch ext {
	case ".mp3":
		w.Header().Set("Content-Type", "audio/mpeg")
	case ".flac":
		w.Header().Set("Content-Type", "audio/flac")
	case ".ogg":
		w.Header().Set("Content-Type", "audio/ogg")
	case ".wav":
		w.Header().Set("Content-Type", "audio/wav")
	case ".jpg", ".jpeg":
		w.Header().Set("Content-Type", "image/jpeg")
	case ".png":
		w.Header().Set("Content-Type", "image/png")
	case ".txt":
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	case ".yaml", ".yml":
		w.Header().Set("Content-Type", "text/yaml; charset=utf-8")
	}

	// Serve file
	file, err := os.Open(fullPath)
	if err != nil {
		http.Error(w, "Failed to open file", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	io.Copy(w, file)
}

func (s *PreviewServer) handleManifest(w http.ResponseWriter, r *http.Request) {
	manifestPath := filepath.Join(s.bundlePath, "manifest.yaml")
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		http.Error(w, "Failed to read manifest", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/yaml; charset=utf-8")
	w.Write(data)
}

const previewHTML = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.Release.Title}} - {{.Release.Artist}} | Rice Preview</title>
    <style>
        * {
            box-sizing: border-box;
            margin: 0;
            padding: 0;
        }
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            background: linear-gradient(135deg, #1a1a2e 0%, #16213e 100%);
            min-height: 100vh;
            color: #e0e0e0;
            padding: 40px 20px;
        }
        .container {
            max-width: 800px;
            margin: 0 auto;
        }
        .album-header {
            display: flex;
            gap: 30px;
            margin-bottom: 40px;
        }
        .cover {
            width: 250px;
            height: 250px;
            border-radius: 8px;
            box-shadow: 0 10px 40px rgba(0,0,0,0.4);
            object-fit: cover;
        }
        .album-info {
            flex: 1;
            display: flex;
            flex-direction: column;
            justify-content: center;
        }
        .album-title {
            font-size: 2rem;
            font-weight: 700;
            margin-bottom: 8px;
            color: #fff;
        }
        .album-artist {
            font-size: 1.2rem;
            color: #aaa;
            margin-bottom: 16px;
        }
        .album-meta {
            font-size: 0.9rem;
            color: #888;
        }
        .tracks {
            background: rgba(255,255,255,0.05);
            border-radius: 8px;
            overflow: hidden;
        }
        .track {
            display: flex;
            align-items: center;
            padding: 16px 20px;
            border-bottom: 1px solid rgba(255,255,255,0.05);
            transition: background 0.2s;
        }
        .track:hover {
            background: rgba(255,255,255,0.05);
        }
        .track:last-child {
            border-bottom: none;
        }
        .track-number {
            width: 30px;
            color: #888;
            font-size: 0.9rem;
        }
        .track-title {
            flex: 1;
            font-weight: 500;
        }
        .track-duration {
            color: #888;
            font-size: 0.9rem;
            margin-right: 20px;
        }
        .play-btn {
            background: #4ade80;
            border: none;
            border-radius: 50%;
            width: 36px;
            height: 36px;
            cursor: pointer;
            display: flex;
            align-items: center;
            justify-content: center;
            transition: transform 0.2s, background 0.2s;
        }
        .play-btn:hover {
            transform: scale(1.1);
            background: #22c55e;
        }
        .play-btn svg {
            width: 14px;
            height: 14px;
            fill: #1a1a2e;
            margin-left: 2px;
        }
        .now-playing {
            margin-top: 40px;
            padding: 20px;
            background: rgba(255,255,255,0.05);
            border-radius: 8px;
        }
        .now-playing h3 {
            font-size: 0.9rem;
            color: #888;
            text-transform: uppercase;
            letter-spacing: 1px;
            margin-bottom: 10px;
        }
        audio {
            width: 100%;
            margin-top: 10px;
        }
        .badge {
            display: inline-block;
            padding: 4px 8px;
            background: rgba(74,222,128,0.2);
            color: #4ade80;
            border-radius: 4px;
            font-size: 0.75rem;
            margin-right: 8px;
        }
        .footer {
            margin-top: 40px;
            text-align: center;
            color: #666;
            font-size: 0.85rem;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="album-header">
            <img src="/files/images/{{.Images.Cover.Filename}}" alt="Album Cover" class="cover">
            <div class="album-info">
                <h1 class="album-title">{{.Release.Title}}</h1>
                <div class="album-artist">{{.Release.Artist}}</div>
                <div class="album-meta">
                    <span class="badge">{{.Release.Genre}}</span>
                    {{if .Release.Subgenre}}<span class="badge">{{.Release.Subgenre}}</span>{{end}}
                    <br><br>
                    Released: {{.Release.ReleaseDate}}<br>
                    {{len .Tracks}} tracks
                </div>
            </div>
        </div>

        <div class="tracks">
            {{range .Tracks}}
            <div class="track">
                <div class="track-number">{{.Number}}</div>
                <div class="track-title">{{.Title}}</div>
                <div class="track-duration">{{.Duration}}</div>
                <button class="play-btn" onclick="playTrack('{{.Filename}}', '{{.Title}}')">
                    <svg viewBox="0 0 24 24"><path d="M8 5v14l11-7z"/></svg>
                </button>
            </div>
            {{end}}
        </div>

        <div class="now-playing" id="now-playing" style="display: none;">
            <h3>Now Playing</h3>
            <div id="now-playing-title"></div>
            <audio id="audio-player" controls></audio>
        </div>

        <div class="footer">
            &copy; {{.Rights.CopyrightYear}} {{.Rights.CopyrightHolder}}<br>
            <small>Preview powered by Rice CLI</small>
        </div>
    </div>

    <script>
        const audioFormats = [{{range .AudioFormats}}'{{.Format}}',{{end}}];
        const preferredFormat = audioFormats.includes('mp3') ? 'mp3' : audioFormats[0];

        function playTrack(filename, title) {
            const player = document.getElementById('audio-player');
            const nowPlaying = document.getElementById('now-playing');
            const nowPlayingTitle = document.getElementById('now-playing-title');

            player.src = '/files/audio/' + filename + '.' + preferredFormat;
            nowPlayingTitle.textContent = title;
            nowPlaying.style.display = 'block';
            player.play();
        }
    </script>
</body>
</html>`
