package pipeline

import (
	"context"

	"github.com/aniketwaliyan/etl-framework/internal/utils/config"
)

type DataRecord map[string]interface{}

type Extractor interface {
	Init(ctx context.Context, cfg *config.PipelineConfig) error
	Extract(ctx context.Context) (<-chan DataRecord, <-chan error)
	Close() error
}

type Transformer interface {
	Init(ctx context.Context, cfg *config.PipelineConfig) error
	Transform(ctx context.Context, input <-chan DataRecord) (<-chan DataRecord, <-chan error)
	Close() error
}

type Loader interface {
	Init(ctx context.Context, cfg *config.PipelineConfig) error
	Load(ctx context.Context, input <-chan DataRecord) error
	Close() error
}

type Pipeline interface {
	Init(ctx context.Context, cfg *config.PipelineConfig) error
	Run(ctx context.Context) error
	Close() error
}
