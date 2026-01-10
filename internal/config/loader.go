package config

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/maxcelant/sinkplot/internal/admission"
	"github.com/maxcelant/sinkplot/internal/schema"
	"gopkg.in/yaml.v3"
)

func Load(path string) (cfg *schema.Config, err error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, fmt.Errorf("config file does not exist: %s", path)
	}

	log.Printf("loading initial config from '%s'", path)
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

	if err := admission.Default(&cfg.App); err != nil {
		return nil, fmt.Errorf("failed to default config object: %w", err)
	}
	if err := admission.Validate(&cfg.App); err != nil {
		return nil, fmt.Errorf("failed to validate config object: %w", err)
	}

	return
}
