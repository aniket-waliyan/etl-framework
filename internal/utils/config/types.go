package config

import "time"

type PipelineConfig struct {
	Pipeline struct {
		Name        string        `yaml:"name"`
		Description string        `yaml:"description"`
		Schedule    string        `yaml:"schedule"`
		Retries     int           `yaml:"retries"`
		RetryDelay  time.Duration `yaml:"retry_delay"`
	} `yaml:"pipeline"`

	Source struct {
		Type     string   `yaml:"type"`
		Servers  []string `yaml:"servers"`
		Database string   `yaml:"database"`
		Table    string   `yaml:"table"`
	} `yaml:"source"`

	Sink struct {
		Type     string `yaml:"type"`
		Server   string `yaml:"server"`
		Database string `yaml:"database"`
		Table    string `yaml:"table"`
	} `yaml:"sink"`

	Transformations []TransformationConfig `yaml:"transformations"`
}

type TransformationConfig struct {
	Type         string      `yaml:"type"`
	ColumnName   string      `yaml:"column_name,omitempty"`
	DefaultValue interface{} `yaml:"default_value,omitempty"`
}
