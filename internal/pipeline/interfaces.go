package pipeline

import (
	"context"

	"github.com/aniketwaliyan/etl-framework/internal/utils/config"
)

// DataRecord represents a single record in the pipeline
type DataRecord map[string]interface{}

// Extractor defines the interface for data extraction
type Extractor interface {
	Init(ctx context.Context, cfg *config.PipelineConfig) error
	Extract(ctx context.Context) (<-chan DataRecord, <-chan error)
	Close() error
}

// Transformer defines the interface for data transformation
type Transformer interface {
	Init(ctx context.Context, cfg *config.PipelineConfig) error
	Transform(ctx context.Context, input <-chan DataRecord) (<-chan DataRecord, <-chan error)
	Close() error
}

// Loader defines the interface for data loading
type Loader interface {
	Init(ctx context.Context, cfg *config.PipelineConfig) error
	Load(ctx context.Context, input <-chan DataRecord) error
	Close() error
}

// Pipeline defines the interface for a complete ETL pipeline
type Pipeline interface {
	Init(ctx context.Context, cfg *config.PipelineConfig) error
	Run(ctx context.Context) error
	Close() error
}
