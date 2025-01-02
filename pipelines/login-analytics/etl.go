package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/aniketwaliyan/etl-framework/pkg/config"
	"github.com/aniketwaliyan/etl-framework/pkg/env"
	"github.com/aniketwaliyan/etl-framework/pkg/pipeline"
	_ "github.com/denisenkom/go-mssqldb"
	_ "github.com/lib/pq"
)

// UserConnectionData represents a record in the source/sink tables
type UserConnectionData struct {
	DealerID         string
	GroupID          string
	DealerCode       string
	LogonLogoffTime  int64
	LoginAllowed     int
	SuccessFailure   int16
	LogonLogoffFlag  string
	Details          string
	ModeOfConnection int
	ConnectionNumber int
	EntrySequence    int
	OMSSequenceNo    int64
	SessionID        string
	SourceTable      string
	ProcessedAt      time.Time
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
		log.Printf("Connecting to SQL Server shard: %s", server)
		parts := strings.Split(server, ":")
		if len(parts) != 2 {
			return fmt.Errorf("invalid server format %s, expected host:port", server)
		}
		host, port := parts[0], parts[1]

		// Build connection string to match the working Node.js configuration
		connStr := fmt.Sprintf("server=%s;port=%s;user id=%s;password=%s;database=%s;"+
			"encrypt=false;trustServerCertificate=true",
			host, port, "TECHCONN", "T$CHC@Nn#2025", "NSEBSE")

		log.Printf("Attempting to connect to SQL Server with user: TECHCONN, database: NSEBSE")

		db, err := sql.Open("sqlserver", connStr)
		if err != nil {
			return fmt.Errorf("failed to connect to SQL Server shard %s: %v", server, err)
		}

		// Set connection pool settings
		db.SetMaxOpenConns(10)
		db.SetMaxIdleConns(5)
		db.SetConnMaxLifetime(time.Minute * 5)

		// Test connection with retry
		var connected bool
		var lastErr error
		for retries := 0; retries < 3; retries++ {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			err := db.PingContext(ctx)
			cancel()

			if err != nil {
				lastErr = err
				log.Printf("Failed to ping SQL Server shard %s (attempt %d/3): %v", server, retries+1, err)
				if retries < 2 {
					time.Sleep(time.Second * 2)
					continue
				}
			} else {
				connected = true
				break
			}
		}

		if !connected {
			db.Close()
			return fmt.Errorf("failed to connect to SQL Server shard %s: %v", server, lastErr)
		}

		log.Printf("Successfully connected to SQL Server shard: %s", server)
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

		var wg sync.WaitGroup
		for i, db := range e.dbs {
			wg.Add(2) // 2 tables per shard

			// Extract from UserConnectionHistory
			go func(shardID int, db *sql.DB) {
				defer wg.Done()
				if err := e.extractTable(ctx, db, shardID, "dbo.tbl_UserConnectionHistory", dataCh, errCh); err != nil {
					errCh <- fmt.Errorf("failed to extract from UserConnectionHistory on shard %d: %v", shardID, err)
				}
			}(i+1, db)

			// Extract from UserConnectionLog
			go func(shardID int, db *sql.DB) {
				defer wg.Done()
				if err := e.extractTable(ctx, db, shardID, "dbo.tbl_UserConnectionLog", dataCh, errCh); err != nil {
					errCh <- fmt.Errorf("failed to extract from UserConnectionLog on shard %d: %v", shardID, err)
				}
			}(i+1, db)
		}

		wg.Wait()
	}()

	return dataCh, errCh
}

func (e *Extractor) extractTable(ctx context.Context, db *sql.DB, shardID int, tableName string, dataCh chan<- pipeline.DataRecord, errCh chan<- error) error {
	query := fmt.Sprintf(`
		SELECT 
			sDealerId,
			sGroupId,
			sDealerCode,
			nLogonLogoffTime,
			nLoginAllowed,
			nSuccessFailure,
			cLogonLogoffFlag,
			sDetails,
			nModeOfConnection,
			nConnectioNumber,
			nEntrySequence,
			nOMSSequenceNo,
			sSessionId
		FROM %s
		WHERE DATEADD(SECOND, nLogonLogoffTime, '1970-01-01') >= DATEADD(HOUR, -5, GETDATE())
		ORDER BY nLogonLogoffTime DESC`, tableName)

	// Execute the query
	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to query table %s on shard %d: %v", tableName, shardID, err)
	}
	defer rows.Close()

	recordCount := 0
	for rows.Next() {
		var data UserConnectionData
		err := rows.Scan(
			&data.DealerID,
			&data.GroupID,
			&data.DealerCode,
			&data.LogonLogoffTime,
			&data.LoginAllowed,
			&data.SuccessFailure,
			&data.LogonLogoffFlag,
			&data.Details,
			&data.ModeOfConnection,
			&data.ConnectionNumber,
			&data.EntrySequence,
			&data.OMSSequenceNo,
			&data.SessionID,
		)
		if err != nil {
			errCh <- fmt.Errorf("failed to scan row from %s on shard %d: %v", tableName, shardID, err)
			continue
		}

		data.SourceTable = fmt.Sprintf("%s_shard%d", tableName, shardID)
		data.ProcessedAt = time.Now()

		recordCount++
		select {
		case <-ctx.Done():
			return ctx.Err()
		case dataCh <- pipeline.DataRecord{Data: data}:
		}
	}
	log.Printf("Extracted %d records from %s on shard %d", recordCount, tableName, shardID)
	return nil
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
	record, ok := data.(UserConnectionData)
	if !ok {
		return nil, fmt.Errorf("invalid data type: expected UserConnectionData")
	}

	// Add any necessary transformations here
	// For now, we're just passing through the data
	return record, nil
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
		l.env.PostgresHost,
		l.env.PostgresPort,
		l.env.PostgresUser,
		l.env.PostgresPassword,
		l.env.PostgresDB)

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
	record, ok := data.(UserConnectionData)
	if !ok {
		l.errors++
		return fmt.Errorf("invalid data type: expected UserConnectionData")
	}

	// Determine target table based on source
	targetTable := "user_connection_log"
	if record.SourceTable == "dbo.tbl_UserConnectionHistory" {
		targetTable = "user_connection_history"
	}

	query := fmt.Sprintf(`
		INSERT INTO %s (
			dealer_id,
			group_id,
			dealer_code,
			logon_logoff_time,
			login_allowed,
			success_failure,
			logon_logoff_flag,
			details,
			mode_of_connection,
			connection_number,
			entry_sequence,
			oms_sequence_no,
			session_id,
			processed_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		ON CONFLICT (dealer_id, logon_logoff_time, entry_sequence) 
		DO UPDATE SET
			group_id = EXCLUDED.group_id,
			dealer_code = EXCLUDED.dealer_code,
			login_allowed = EXCLUDED.login_allowed,
			success_failure = EXCLUDED.success_failure,
			logon_logoff_flag = EXCLUDED.logon_logoff_flag,
			details = EXCLUDED.details,
			mode_of_connection = EXCLUDED.mode_of_connection,
			connection_number = EXCLUDED.connection_number,
			oms_sequence_no = EXCLUDED.oms_sequence_no,
			session_id = EXCLUDED.session_id,
			processed_at = EXCLUDED.processed_at`, targetTable)

	_, err := l.db.ExecContext(ctx, query,
		record.DealerID,
		record.GroupID,
		record.DealerCode,
		record.LogonLogoffTime,
		record.LoginAllowed,
		record.SuccessFailure,
		record.LogonLogoffFlag,
		record.Details,
		record.ModeOfConnection,
		record.ConnectionNumber,
		record.EntrySequence,
		record.OMSSequenceNo,
		record.SessionID,
		record.ProcessedAt,
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
