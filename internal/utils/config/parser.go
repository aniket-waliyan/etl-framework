package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Parser struct{}

func NewParser() *Parser {
	return &Parser{}
}

func (p *Parser) Parse(filePath string) (*PipelineConfig, error) {

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("configuration file not found: %s", filePath)
	}

	data, err := os.ReadFile(filepath.Clean(filePath))
	if err != nil {
		return nil, fmt.Errorf("error reading configuration file: %w", err)
	}

	var config PipelineConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("error parsing configuration: %w", err)
	}

	if err := p.validate(&config); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &config, nil
}

func (p *Parser) validate(config *PipelineConfig) error {
	if config.Pipeline.Name == "" {
		return fmt.Errorf("pipeline name is required")
	}
	if config.Pipeline.Retries < 0 {
		return fmt.Errorf("retries must be non-negative")
	}
	if config.Source.Type == "" {
		return fmt.Errorf("source type is required")
	}
	if config.Sink.Type == "" {
		return fmt.Errorf("sink type is required")
	}
	return nil
}
