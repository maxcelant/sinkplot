package schema

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

func Load(path string) (cfg *Config, err error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, fmt.Errorf("config file does not exist: %s", path)
	}

	buf, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	ext := filepath.Ext(path)
	switch ext {
	case ".json":
		err = json.Unmarshal(buf, &cfg)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal config from json: %w", err)
		}
	case ".yaml", ".yml":
		err = yaml.Unmarshal(buf, &cfg)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal config from yaml: %w", err)
		}
	default:
		return nil, fmt.Errorf("unsupported config file format: %s (expected .json, .yaml, or .yml)", ext)
	}

	return
}
