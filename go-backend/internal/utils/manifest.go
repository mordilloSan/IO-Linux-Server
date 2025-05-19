package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

func ParseViteManifest(manifestPath string) (string, string, error) {
	raw, err := os.ReadFile(manifestPath)
	if err != nil {
		return "", "", fmt.Errorf("failed to read manifest: %w", err)
	}

	var manifest map[string]ManifestEntry
	if err := json.Unmarshal(raw, &manifest); err != nil {
		return "", "", fmt.Errorf("failed to parse manifest: %w", err)
	}

	entry, ok := manifest["index.html"]
	if !ok {
		return "", "", fmt.Errorf("entry not found in manifest")
	}

	js := filepath.ToSlash("/" + entry.File)
	css := ""
	if len(entry.CSS) > 0 {
		css = filepath.ToSlash("/" + entry.CSS[0])
	}

	return js, css, nil
}

func ParseViteManifestBytes(data []byte) (js string, css string, err error) {
	var manifest ViteManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return "", "", fmt.Errorf("failed to parse manifest: %w", err)
	}

	// Example: Find entrypoint (usually "src/main.tsx" or similar)
	for _, entry := range manifest {
		js = entry.File
		if len(entry.Css) > 0 {
			css = entry.Css[0]
		}
		break // only first, adjust as needed!
	}
	return js, css, nil
}
