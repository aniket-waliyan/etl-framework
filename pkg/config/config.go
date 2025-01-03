package config

import (
	"fmt"
	"os"
	"regexp"
	"strconv"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Pipeline struct {
		Name        string `yaml:"name"`
		Description string `yaml:"description"`
		Schedule    string `yaml:"schedule"`
		Retries     int    `yaml:"retries"`
		RetryDelay  string `yaml:"retry_delay"`
	} `yaml:"pipeline"`
	Source struct {
		Type     string `yaml:"type"`
		Database string `yaml:"database"`
		Table    string `yaml:"table"`
	} `yaml:"source"`
	Sink struct {
		Type  string `yaml:"type"`
		Table string `yaml:"table"`
	} `yaml:"sink"`
}

type Parser struct{}

func NewParser() *Parser {
	return &Parser{}
}

func (p *Parser) Parse(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %v", err)
	}

	content := string(data)
	envVarPattern := regexp.MustCompile(`\${([^}]+)}`)
	content = envVarPattern.ReplaceAllStringFunc(content, func(match string) string {
		envVar := match[2 : len(match)-1]
		value := os.Getenv(envVar)
		if value == "" {

			return match
		}
		return value
	})

	var cfg Config
	if err := yaml.Unmarshal([]byte(content), &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %v", err)
	}

	return &cfg, nil
}

func getIntOrDefault(value string, defaultValue int) int {
	if i, err := strconv.Atoi(value); err == nil {
		return i
	}
	return defaultValue
}
