package pipeline

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aniketwaliyan/etl-framework/internal/utils/config"
)

// Orchestrator manages the execution of ETL pipelines
type Orchestrator struct {
	config      *config.PipelineConfig
	extractor   Extractor
	transformer Transformer
	loader      Loader
}

// NewOrchestrator creates a new pipeline orchestrator
func NewOrchestrator(cfg *config.PipelineConfig, ext Extractor, trans Transformer, load Loader) *Orchestrator {
	return &Orchestrator{
		config:      cfg,
		extractor:   ext,
		transformer: trans,
		loader:      load,
	}
}

// Execute runs the pipeline with retry logic
func (o *Orchestrator) Execute(ctx context.Context) error {
	var lastErr error

	for attempt := 0; attempt <= o.config.Pipeline.Retries; attempt++ {
		if attempt > 0 {
			log.Printf("Retry attempt %d/%d after error: %v",
				attempt, o.config.Pipeline.Retries, lastErr)
			time.Sleep(o.config.Pipeline.RetryDelay)
		}

		if err := o.runPipeline(ctx); err != nil {
			lastErr = err
			continue
		}

		return nil // Success
	}

	return fmt.Errorf("pipeline failed after %d retries, last error: %v",
		o.config.Pipeline.Retries, lastErr)
}

// runPipeline executes a single pipeline run
func (o *Orchestrator) runPipeline(ctx context.Context) error {
	// Initialize components
	if err := o.initComponents(ctx); err != nil {
		return fmt.Errorf("initialization failed: %w", err)
	}
	defer o.closeComponents()

	// Start extraction
	records, extractErrs := o.extractor.Extract(ctx)

	// Apply transformations
	transformed, transformErrs := o.transformer.Transform(ctx, records)

	// Handle loading and errors
	errCh := make(chan error, 1)
	go func() {
		defer close(errCh)
		if err := o.loader.Load(ctx, transformed); err != nil {
			errCh <- err
		}
	}()

	// Error handling
	for {
		select {
		case err, ok := <-extractErrs:
			if !ok {
				extractErrs = nil
				continue
			}
			return fmt.Errorf("extraction error: %w", err)

		case err, ok := <-transformErrs:
			if !ok {
				transformErrs = nil
				continue
			}
			return fmt.Errorf("transformation error: %w", err)

		case err, ok := <-errCh:
			if !ok {
				return nil
			}
			return fmt.Errorf("loading error: %w", err)

		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (o *Orchestrator) initComponents(ctx context.Context) error {
	if err := o.extractor.Init(ctx, o.config); err != nil {
		return fmt.Errorf("extractor initialization failed: %w", err)
	}
	if err := o.transformer.Init(ctx, o.config); err != nil {
		return fmt.Errorf("transformer initialization failed: %w", err)
	}
	if err := o.loader.Init(ctx, o.config); err != nil {
		return fmt.Errorf("loader initialization failed: %w", err)
	}
	return nil
}

func (o *Orchestrator) closeComponents() {
	if err := o.extractor.Close(); err != nil {
		log.Printf("Error closing extractor: %v", err)
	}
	if err := o.transformer.Close(); err != nil {
		log.Printf("Error closing transformer: %v", err)
	}
	if err := o.loader.Close(); err != nil {
		log.Printf("Error closing loader: %v", err)
	}
}
