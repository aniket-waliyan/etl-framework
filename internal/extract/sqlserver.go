package extract

import (
	"context"
	"database/sql"
	"fmt"
	"sync"

	"github.com/aniketwaliyan/etl-framework/internal/pipeline"
	"github.com/aniketwaliyan/etl-framework/internal/utils/config"
	_ "github.com/denisenkom/go-mssqldb"
)

type SQLServerExtractor struct {
	config *config.PipelineConfig
	dbs    []*sql.DB
}

func NewSQLServerExtractor() *SQLServerExtractor {
	return &SQLServerExtractor{}
}

func (e *SQLServerExtractor) Init(ctx context.Context, cfg *config.PipelineConfig) error {
	e.config = cfg
	e.dbs = make([]*sql.DB, len(cfg.Source.Servers))

	for i, server := range cfg.Source.Servers {
		connStr := fmt.Sprintf("server=%s;database=%s;", server, cfg.Source.Database)
		db, err := sql.Open("sqlserver", connStr)
		if err != nil {
			return fmt.Errorf("failed to connect to server %s: %w", server, err)
		}
		e.dbs[i] = db
	}

	return nil
}

func (e *SQLServerExtractor) Extract(ctx context.Context) (<-chan pipeline.DataRecord, <-chan error) {
	records := make(chan pipeline.DataRecord)
	errs := make(chan error, 1)
	var wg sync.WaitGroup

	for _, db := range e.dbs {
		wg.Add(1)
		go func(db *sql.DB) {
			defer wg.Done()
			if err := e.extractFromDB(ctx, db, records); err != nil {
				errs <- err
			}
		}(db)
	}

	go func() {
		wg.Wait()
		close(records)
		close(errs)
	}()

	return records, errs
}

func (e *SQLServerExtractor) extractFromDB(ctx context.Context, db *sql.DB, records chan<- pipeline.DataRecord) error {
	query := fmt.Sprintf("SELECT * FROM %s", e.config.Source.Table)
	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return fmt.Errorf("failed to get columns: %w", err)
	}

	for rows.Next() {
		values := make([]interface{}, len(cols))
		valuePtrs := make([]interface{}, len(cols))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return fmt.Errorf("scan failed: %w", err)
		}

		record := make(pipeline.DataRecord)
		for i, col := range cols {
			record[col] = values[i]
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case records <- record:
		}
	}

	return rows.Err()
}

func (e *SQLServerExtractor) Close() error {
	var errs []error
	for _, db := range e.dbs {
		if err := db.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("failed to close some connections: %v", errs)
	}
	return nil
}
