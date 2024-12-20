package pipeline

import (
	"context"

	"github.com/aniketwaliyan/etl-framework/internal/utils/config"
)

type NoopTransformer struct{}

func (t *NoopTransformer) Init(ctx context.Context, cfg *config.PipelineConfig) error { return nil }

func (t *NoopTransformer) Transform(ctx context.Context, input <-chan DataRecord) (<-chan DataRecord, <-chan error) {
	return input, make(chan error)
}

func (t *NoopTransformer) Close() error { return nil }

type NoopLoader struct{}

func (l *NoopLoader) Init(ctx context.Context, cfg *config.PipelineConfig) error { return nil }

func (l *NoopLoader) Load(ctx context.Context, input <-chan DataRecord) error { return nil }

func (l *NoopLoader) Close() error { return nil }
