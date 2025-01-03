package pipeline

import (
	"context"
	"fmt"

	"github.com/aniketwaliyan/etl-framework/pkg/config"
)

type DataRecord struct {
	Data interface{}
}

type Extractor interface {
	Init(*config.Config) error
	Extract(context.Context) (<-chan DataRecord, <-chan error)
	Close() error
}

type Transformer interface {
	Init(*config.Config) error
	Transform(context.Context, interface{}) (interface{}, error)
	Close() error
}

type Loader interface {
	Init(*config.Config) error
	Load(context.Context, interface{}) error
	Close() error
}

type Orchestrator struct {
	config      *config.Config
	extractor   Extractor
	transformer Transformer
	loader      Loader
}

func NewOrchestrator(cfg *config.Config, e Extractor, t Transformer, l Loader) *Orchestrator {
	return &Orchestrator{
		config:      cfg,
		extractor:   e,
		transformer: t,
		loader:      l,
	}
}

func (o *Orchestrator) Execute(ctx context.Context) error {

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


	dataCh, errCh := o.extractor.Extract(ctx)


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

		
			transformed, err := o.transformer.Transform(ctx, record.Data)
			if err != nil {
				return fmt.Errorf("transformation error: %v", err)
			}

		
			if err := o.loader.Load(ctx, transformed); err != nil {
				return fmt.Errorf("loading error: %v", err)
			}
		}
	}
}
