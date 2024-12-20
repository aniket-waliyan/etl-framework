package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/aniketwaliyan/etl-framework/pkg/config"
	"github.com/aniketwaliyan/etl-framework/pkg/env"
	"github.com/aniketwaliyan/etl-framework/pkg/pipeline"
	_ "github.com/denisenkom/go-mssqldb"
	_ "github.com/lib/pq"
)

// CustomerData represents a record in the source/sink tables
type CustomerData struct {
	CustomerID int
	ShardID    int
	FirstName  string
	LastName   string
	Email      string
	CreatedAt  time.Time
	Country    string
	Amount     float64
}

// Extractor handles data extraction from SQL Server shards
type Extractor struct {
	dbs []*sql.DB
	env *env.Config
}

func NewExtractor(envConfig *env.Config) *Extractor {
	return &Extractor{env: envConfig}
}

func (e *Extractor) Init(cfg *config.Config) error {
	log.Println("Initializing Extractor...")
	for _, server := range e.env.SQLServerShards {
		log.Printf("Connecting to SQL Server: %s", server)
		parts := strings.Split(server, ":")
		if len(parts) != 2 {
			return fmt.Errorf("invalid server format %s, expected host:port", server)
		}
		host, port := parts[0], parts[1]

		connStr := fmt.Sprintf("server=%s,%s;user id=%s;password=%s;database=%s;encrypt=disable",
			host, port, e.env.SQLServerUser, e.env.SQLServerPassword, e.env.SQLServerDB)
		db, err := sql.Open("sqlserver", connStr)
		if err != nil {
			return fmt.Errorf("failed to connect to SQL Server %s: %v", server, err)
		}
		if err := db.Ping(); err != nil {
			return fmt.Errorf("failed to ping SQL Server %s: %v", server, err)
		}
		log.Printf("Successfully connected to SQL Server: %s", server)
		e.dbs = append(e.dbs, db)
	}
	return nil
}

func (e *Extractor) Extract(ctx context.Context) (<-chan pipeline.DataRecord, <-chan error) {
	dataCh := make(chan pipeline.DataRecord)
	errCh := make(chan error)

	go func() {
		defer close(dataCh)
		defer close(errCh)

		for i, db := range e.dbs {
			log.Printf("Extracting data from shard %d", i+1)
			query := `SELECT customer_id, shard_id, first_name, last_name, 
				     email, created_at, country, amount FROM source_table`

			rows, err := db.QueryContext(ctx, query)
			if err != nil {
				errCh <- fmt.Errorf("failed to query source table: %v", err)
				return
			}
			defer rows.Close()

			recordCount := 0
			for rows.Next() {
				var data CustomerData
				err := rows.Scan(
					&data.CustomerID,
					&data.ShardID,
					&data.FirstName,
					&data.LastName,
					&data.Email,
					&data.CreatedAt,
					&data.Country,
					&data.Amount,
				)
				if err != nil {
					errCh <- fmt.Errorf("failed to scan row: %v", err)
					return
				}

				recordCount++
				select {
				case <-ctx.Done():
					return
				case dataCh <- pipeline.DataRecord{Data: data}:
				}
			}
			log.Printf("Extracted %d records from shard %d", recordCount, i+1)
		}
		log.Println("Extraction completed")
	}()

	return dataCh, errCh
}

func (e *Extractor) Close() error {
	for _, db := range e.dbs {
		if err := db.Close(); err != nil {
			return err
		}
	}
	return nil
}

// Transformer handles data transformation
type Transformer struct{}

func NewTransformer() *Transformer {
	return &Transformer{}
}

func (t *Transformer) Init(cfg *config.Config) error {
	return nil
}

func (t *Transformer) Transform(ctx context.Context, data interface{}) (interface{}, error) {
	// In this case, we're not modifying the data, just passing it through
	// You could add data validation, enrichment, or transformation here
	return data, nil
}

func (t *Transformer) Close() error {
	return nil
}

// Loader handles data loading to PostgreSQL
type Loader struct {
	db     *sql.DB
	count  int
	errors int
	env    *env.Config
}

func NewLoader(envConfig *env.Config) *Loader {
	return &Loader{env: envConfig}
}

func (l *Loader) Init(cfg *config.Config) error {
	log.Println("Initializing Loader...")
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		l.env.PostgresHost, l.env.PostgresPort, l.env.PostgresUser, l.env.PostgresPassword, l.env.PostgresDB)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return fmt.Errorf("failed to connect to PostgreSQL: %v", err)
	}
	if err := db.Ping(); err != nil {
		return fmt.Errorf("failed to ping PostgreSQL: %v", err)
	}
	log.Println("Successfully connected to PostgreSQL")
	l.db = db
	return nil
}

func (l *Loader) Load(ctx context.Context, data interface{}) error {
	customer, ok := data.(CustomerData)
	if !ok {
		l.errors++
		return fmt.Errorf("invalid data type: expected CustomerData")
	}

	query := `INSERT INTO sink_table (
		customer_id, shard_id, first_name, last_name, 
		email, created_at, country, amount, processed_at
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, CURRENT_TIMESTAMP)
	ON CONFLICT (customer_id) DO UPDATE SET
		shard_id = EXCLUDED.shard_id,
		first_name = EXCLUDED.first_name,
		last_name = EXCLUDED.last_name,
		email = EXCLUDED.email,
		created_at = EXCLUDED.created_at,
		country = EXCLUDED.country,
		amount = EXCLUDED.amount,
		processed_at = CURRENT_TIMESTAMP`

	_, err := l.db.ExecContext(ctx, query,
		customer.CustomerID,
		customer.ShardID,
		customer.FirstName,
		customer.LastName,
		customer.Email,
		customer.CreatedAt,
		customer.Country,
		customer.Amount,
	)

	if err != nil {
		l.errors++
		return fmt.Errorf("failed to insert/update record: %v", err)
	}

	l.count++
	if l.count%100 == 0 {
		log.Printf("Loaded %d records (errors: %d)", l.count, l.errors)
	}

	return nil
}

func (l *Loader) Close() error {
	log.Printf("ETL completed. Total records loaded: %d, Total errors: %d", l.count, l.errors)
	if l.db != nil {
		return l.db.Close()
	}
	return nil
}

func main() {
	log.Println("Starting ETL pipeline...")
	ctx := context.Background()

	// Load environment variables
	workDir, err := os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get working directory: %v", err)
	}
	envConfig, err := env.Load(workDir)
	if err != nil {
		log.Fatalf("Failed to load environment variables: %v", err)
	}

	// Parse configuration
	log.Println("Parsing configuration...")
	parser := config.NewParser()
	cfg, err := parser.Parse("config.yaml")
	if err != nil {
		log.Fatalf("Failed to parse configuration: %v", err)
	}

	// Initialize components
	log.Println("Initializing pipeline components...")
	extractor := NewExtractor(envConfig)
	if err := extractor.Init(cfg); err != nil {
		log.Fatalf("Failed to initialize extractor: %v", err)
	}

	transformer := NewTransformer()
	if err := transformer.Init(cfg); err != nil {
		log.Fatalf("Failed to initialize transformer: %v", err)
	}

	loader := NewLoader(envConfig)
	if err := loader.Init(cfg); err != nil {
		log.Fatalf("Failed to initialize loader: %v", err)
	}

	// Create and run pipeline
	log.Println("Starting pipeline execution...")
	orchestrator := pipeline.NewOrchestrator(cfg, extractor, transformer, loader)
	if err := orchestrator.Execute(ctx); err != nil {
		log.Fatalf("Pipeline execution failed: %v", err)
	}
	log.Println("Pipeline execution completed successfully")
}
