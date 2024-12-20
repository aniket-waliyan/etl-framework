package pipeline

import (
	"context"
	"fmt"

	"github.com/aniketwaliyan/etl-framework/pkg/config"
)

// DataRecord represents a single record in the pipeline
type DataRecord struct {
	Data interface{}
}

// Extractor interface for data extraction
type Extractor interface {
	Init(*config.Config) error
	Extract(context.Context) (<-chan DataRecord, <-chan error)
	Close() error
}

// Transformer interface for data transformation
type Transformer interface {
	Init(*config.Config) error
	Transform(context.Context, interface{}) (interface{}, error)
	Close() error
}

// Loader interface for data loading
type Loader interface {
	Init(*config.Config) error
	Load(context.Context, interface{}) error
	Close() error
}

// Orchestrator manages the ETL pipeline
type Orchestrator struct {
	config      *config.Config
	extractor   Extractor
	transformer Transformer
	loader      Loader
}

// NewOrchestrator creates a new pipeline orchestrator
func NewOrchestrator(cfg *config.Config, e Extractor, t Transformer, l Loader) *Orchestrator {
	return &Orchestrator{
		config:      cfg,
		extractor:   e,
		transformer: t,
		loader:      l,
	}
}

// Execute runs the pipeline
func (o *Orchestrator) Execute(ctx context.Context) error {
	// Initialize components
	if err := o.extractor.Init(o.config); err != nil {
		return fmt.Errorf("failed to initialize extractor: %v", err)
	}
	defer o.extractor.Close()

	if err := o.transformer.Init(o.config); err != nil {
		return fmt.Errorf("failed to initialize transformer: %v", err)
	}
	defer o.transformer.Close()

	if err := o.loader.Init(o.config); err != nil {
		return fmt.Errorf("failed to initialize loader: %v", err)
	}
	defer o.loader.Close()

	// Start extraction
	dataCh, errCh := o.extractor.Extract(ctx)

	// Process data
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case err := <-errCh:
			if err != nil {
				return fmt.Errorf("extraction error: %v", err)
			}
			return nil
		case record, ok := <-dataCh:
			if !ok {
				return nil
			}

			// Transform
			transformed, err := o.transformer.Transform(ctx, record.Data)
			if err != nil {
				return fmt.Errorf("transformation error: %v", err)
			}

			// Load
			if err := o.loader.Load(ctx, transformed); err != nil {
				return fmt.Errorf("loading error: %v", err)
			}
		}
	}
}
