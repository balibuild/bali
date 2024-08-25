package barrow

import "path/filepath"

type Target struct {
	Name        string   `toml:"name"`
	Description string   `toml:"description,omitempty"`
	Destination string   `toml:"destination,omitempty"`
	GoFlags     []string `toml:"goflags,omitempty"`
	Version     string   `toml:"version,omitempty"`
}

func NewTarget(targetDir string) (*Target, error) {
	file := filepath.Join(targetDir, "bailsrc.toml")
	var t Target
	if err := LoadMetadata(file, &t); err != nil {
		return nil, err
	}
	return &t, nil
}
