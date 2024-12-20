package main

const etlTemplate = `package main

import (
	"context"
	"log"

	"github.com/aniketwaliyan/etl-framework/pkg/config"
	"github.com/aniketwaliyan/etl-framework/pkg/pipeline"
)

// Extractor handles data extraction from source
type Extractor struct{}

func NewExtractor() *Extractor {
	return &Extractor{}
}

func (e *Extractor) Init(cfg *config.Config) error {
	return nil
}

func (e *Extractor) Extract(ctx context.Context) (<-chan pipeline.DataRecord, <-chan error) {
	dataCh := make(chan pipeline.DataRecord)
	errCh := make(chan error)

	go func() {
		defer close(dataCh)
		defer close(errCh)
		// TODO: Implement extraction logic
	}()

	return dataCh, errCh
}

func (e *Extractor) Close() error {
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
	// TODO: Implement transformation logic
	return data, nil
}

func (t *Transformer) Close() error {
	return nil
}

// Loader handles data loading to sink
type Loader struct{}

func NewLoader() *Loader {
	return &Loader{}
}

func (l *Loader) Init(cfg *config.Config) error {
	return nil
}

func (l *Loader) Load(ctx context.Context, data interface{}) error {
	// TODO: Implement loading logic
	return nil
}

func (l *Loader) Close() error {
	return nil
}

func main() {
	ctx := context.Background()

	// Parse configuration
	parser := config.NewParser()
	cfg, err := parser.Parse("config.yaml")
	if err != nil {
		log.Fatalf("Failed to parse configuration: %v", err)
	}

	// Initialize components
	extractor := NewExtractor()
	transformer := NewTransformer()
	loader := NewLoader()

	// Create and run pipeline
	orchestrator := pipeline.NewOrchestrator(cfg, extractor, transformer, loader)
	if err := orchestrator.Execute(ctx); err != nil {
		log.Fatalf("Pipeline execution failed: %v", err)
	}
}`

const configTemplate = `# Pipeline Configuration
name: "{{.Name}}"
description: "ETL pipeline for {{.Name}}"

# Scheduling Configuration
# Examples:
# schedule: "@daily"    # Run once a day at midnight
# schedule: "@hourly"   # Run every hour
# schedule: "0 0 * * *" # Run at midnight (00:00) every day
# schedule: "0 */2 * * *" # Run every 2 hours
# schedule: "0 0 * * MON" # Run at midnight every Monday
# schedule: "0 0 1 * *"   # Run at midnight on the first day of every month
schedule: "@daily"

# Error Handling Configuration
error_handling:
  # Number of times to retry a failed pipeline execution
  max_retries: 3
  
  # Time to wait between retry attempts when pipeline fails
  # Examples: "5s", "1m", "2m30s"
  retry_delay: "5s"
  
  # Maximum time to wait between retries
  max_retry_delay: "15m"
  
  # Whether to use exponential backoff for retries
  exponential_backoff: true
  
  # Stop retrying if error persists after this duration
  # Examples: "1h", "30m", "2h30m"
  retry_timeout: "1h"

# Source Database Configuration (SQL Server)
source:
  type: "sqlserver"
  servers:
    - "${SQLSERVER_SHARD1_HOST}:${SQLSERVER_SHARD1_PORT}"
    - "${SQLSERVER_SHARD2_HOST}:${SQLSERVER_SHARD2_PORT}"
    - "${SQLSERVER_SHARD3_HOST}:${SQLSERVER_SHARD3_PORT}"
    - "${SQLSERVER_SHARD4_HOST}:${SQLSERVER_SHARD4_PORT}"
  database: "${SOURCE_DB_NAME}"
  username: "${SOURCE_DB_USER}"
  password: "${SOURCE_DB_PASSWORD}"
  query: "SELECT * FROM source_table"
  
  # Connection retry settings
  connection:
    max_retries: 3
    retry_delay: "1s"
    timeout: "30s"

# Sink Database Configuration (PostgreSQL)
sink:
  type: "postgres"
  host: "${POSTGRES_HOST}"
  port: "${POSTGRES_PORT}"
  database: "${SINK_DB_NAME}"
  username: "${SINK_DB_USER}"
  password: "${SINK_DB_PASSWORD}"
  table: "sink_table"
  
  # Connection retry settings
  connection:
    max_retries: 3
    retry_delay: "1s"
    timeout: "30s"

# Monitoring and Logging
monitoring:
  log_level: "info"
  metrics_enabled: true
  alert_on_failure: true
  alert_email: "${ALERT_EMAIL}"
  
  # Health check configuration
  health_check:
    enabled: true
    interval: "1m"
    timeout: "10s"`

const dockerfileTemplate = `# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY . .
COPY go.mod go.sum ./
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o pipeline

# Final stage
FROM alpine:latest

WORKDIR /app
COPY --from=builder /app/pipeline .
COPY config.yaml .

# Install required dependencies
RUN apk add --no-cache \
    postgresql-client \
    sqlcmd

ENTRYPOINT ["./pipeline"]`

const readmeTemplate = `# {{.Name}} ETL Pipeline

This pipeline is generated using the ETL Pipeline Framework. It extracts data from SQL Server shards and loads it into a PostgreSQL database.

## Structure

- ` + "`" + `etl.go` + "`" + `: Main pipeline implementation
- ` + "`" + `config.yaml` + "`" + `: Pipeline configuration
- ` + "`" + `Dockerfile` + "`" + `: Container configuration

## Configuration

The pipeline is configured through ` + "`" + `config.yaml` + "`" + `:

- Source (SQL Server Shards):
  - Servers: localhost:1433, 1434, 1435, 1436
  - Database: my_source_db
  - Table: source_table

- Sink (PostgreSQL):
  - Server: localhost:5432
  - Database: my_sink_db
  - Table: sink_table

## Running the Pipeline

### Local Development

1. Install dependencies:
   ` + "```bash" + `
   go mod download
   ` + "```" + `

2. Run the pipeline:
   ` + "```bash" + `
   go run etl.go
   ` + "```" + `

### Using Docker

1. Build the container:
   ` + "```bash" + `
   docker build -t {{.Name}}-etl .
   ` + "```" + `

2. Run the container:
   ` + "```bash" + `
   docker run --network=host {{.Name}}-etl
   ` + "```" + `

## Customization

1. Implement extraction logic in ` + "`" + `Extractor.Extract()` + "`" + `
2. Implement transformation logic in ` + "`" + `Transformer.Transform()` + "`" + `
3. Implement loading logic in ` + "`" + `Loader.Load()` + "`" + `

## Error Handling

The pipeline includes:
- Automatic retries (configured in config.yaml)
- Error logging
- Graceful shutdown

## Monitoring

Monitor the pipeline through:
- Application logs
- Database logs
- Container logs when running in Docker`

const envTemplate = `# Database Configuration - Source (SQL Server Shards)
SQLSERVER_SHARD1_HOST=localhost
SQLSERVER_SHARD1_PORT=1433
SQLSERVER_SHARD2_HOST=localhost
SQLSERVER_SHARD2_PORT=1434
SQLSERVER_SHARD3_HOST=localhost
SQLSERVER_SHARD3_PORT=1435
SQLSERVER_SHARD4_HOST=localhost
SQLSERVER_SHARD4_PORT=1436
SOURCE_DB_NAME=my_source_db
SOURCE_DB_USER=sa
SOURCE_DB_PASSWORD=YourStrong@Passw0rd

# Database Configuration - Sink (PostgreSQL)
POSTGRES_HOST=localhost
POSTGRES_PORT=5432
SINK_DB_NAME=my_sink_db
SINK_DB_USER=etl_user
SINK_DB_PASSWORD=etl_password

# Monitoring Configuration
ALERT_EMAIL=alerts@example.com`
