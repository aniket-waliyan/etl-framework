# ETL Framework

A robust, scalable ETL (Extract, Transform, Load) framework designed for handling data pipelines with support for database sharding, parallel processing, and flexible scheduling.

## Features

- **Modular Pipeline Architecture**: Easily create and manage independent ETL pipelines
- **Database Sharding Support**: Handle distributed data across multiple SQL Server shards
- **Parallel Processing**: Efficient data processing with concurrent execution
- **Flexible Scheduling**: Support for various scheduling patterns (daily, hourly, custom cron)
- **Production-Ready**: Systemd service integration with proper logging and monitoring
- **Docker Support**: Containerized development environment for both source and sink databases
- **CLI Tools**: Comprehensive command-line tools for pipeline management

## Prerequisites

- Go 1.19 or later
- Docker and Docker Compose
- SQL Server command-line tools (`sqlcmd`)
- PostgreSQL client (`psql`)
- Linux environment with systemd

## Quick Start

1. **Clone the Repository**
   ```bash
   git clone <repository-url>
   cd etl-framework
   ```

2. **Install Dependencies**
   ```bash
   # Install SQL Server tools
   curl https://packages.microsoft.com/keys/microsoft.asc | sudo tee /etc/apt/trusted.gpg.d/microsoft.asc
   echo "deb [arch=amd64] https://packages.microsoft.com/ubuntu/22.04/prod jammy main" | sudo tee /etc/apt/sources.list.d/mssql-release.list
   sudo apt-get update
   sudo ACCEPT_EULA=Y apt-get install -y mssql-tools unixodbc-dev
   
   # Install PostgreSQL client
   sudo apt-get install -y postgresql-client
   
   # Add SQL Server tools to PATH
   echo 'export PATH="$PATH:/opt/mssql-tools/bin"' >> ~/.bashrc
   source ~/.bashrc
   ```

3. **Setup Development Environment**
   ```bash
   # For development setup
   ./scripts/setup-dev.sh
   
   # For production setup
   ./scripts/setup-dev.sh --prod
   ```

## Creating a New Pipeline

1. **Generate Pipeline Boilerplate**
   ```bash
   ./etl-cli generate --name my-pipeline
   ```

   This creates:
   - `etl.go`: Pipeline implementation with scheduling logic
   - `config.yaml`: Configuration file with schedule settings
   - `.env`: Environment variables
   - `Dockerfile`: Container configuration
   - `README.md`: Pipeline-specific documentation

2. **Configure Pipeline**
   ```bash
   cd pipelines/my-pipeline
   cp .env.template .env
   # Edit .env with your configuration
   ```

## Scheduling Options

Configure pipeline schedules in `config.yaml`:

```yaml
# Daily at midnight
schedule: "@daily"

# Every hour
schedule: "@hourly"

# Every 2 hours
schedule: "0 */2 * * *"

# Weekly on Monday at midnight
schedule: "0 0 * * MON"

# Monthly on 1st at midnight
schedule: "0 0 1 * *"
```

The schedule determines when your pipeline starts its regular execution.

## Error Handling and Retries

The pipeline includes robust error handling with configurable retry behavior:

```yaml
error_handling:
  # Number of times to retry a failed pipeline execution
  max_retries: 3
  
  # Time to wait between retry attempts when pipeline fails
  retry_delay: "5s"
  
  # Maximum time to wait between retries
  max_retry_delay: "15m"
  
  # Whether to use exponential backoff for retries
  exponential_backoff: true
  
  # Stop retrying if error persists after this duration
  retry_timeout: "1h"
```

Important distinctions:
- **Schedule**: Determines when the pipeline starts its regular execution (e.g., daily at midnight)
- **Retry Delay**: Only applies when a pipeline execution fails and needs to be retried
- **Connection Retries**: Separate settings for database connection attempts

Example scenarios:
1. **Normal Execution**:
   - Schedule: `@daily` (runs at midnight)
   - Next execution: Next day at midnight
   - No retries needed if successful

2. **Failed Execution with Retries**:
   - Pipeline starts at midnight
   - Execution fails
   - First retry: After 5 seconds
   - Second retry: After 10 seconds (with exponential backoff)
   - Third retry: After 20 seconds
   - If still failing: Waits for next scheduled run

3. **Database Connection Issues**:
   - Connection fails
   - Retries every 1 second up to 3 times
   - Independent of main pipeline retry settings

Manage schedules using:
```bash
# Set daily schedule at midnight
./scripts/manage-pipeline.sh schedule my-pipeline daily 00:00

# Run every 2 hours
./scripts/manage-pipeline.sh schedule my-pipeline hourly 2

# Run weekly on Monday at midnight
./scripts/manage-pipeline.sh schedule my-pipeline weekly mon 00:00

# Run monthly on 1st at midnight
./scripts/manage-pipeline.sh schedule my-pipeline monthly 1 00:00

# Remove scheduling
./scripts/manage-pipeline.sh schedule my-pipeline remove
```

## Pipeline Management

Control your ETL pipelines:

```bash
# List all pipelines
./scripts/manage-pipeline.sh list

# Start a pipeline
./scripts/manage-pipeline.sh start my-pipeline

# Stop a pipeline
./scripts/manage-pipeline.sh stop my-pipeline

# View pipeline logs
./scripts/manage-pipeline.sh logs my-pipeline

# View detailed information
./scripts/manage-pipeline.sh info my-pipeline

# Delete pipeline service
./scripts/manage-pipeline.sh delete my-pipeline
```

## Database Configuration

### Source Database (SQL Server)
- 4 shards running on ports 1433-1436
- Default credentials in `.env`:
  ```
  SOURCE_DB_USER=sa
  SOURCE_DB_PASSWORD=YourStrong@Passw0rd
  ```

### Sink Database (PostgreSQL)
- Running on port 5432
- Default credentials in `.env`:
  ```
  SINK_DB_USER=etl_user
  SINK_DB_PASSWORD=etl_password
  ```

## Monitoring and Logging

### Log Locations
- Pipeline logs: `/var/log/etl/<pipeline-name>/`
  - `output.log`: Standard output
  - `error.log`: Error messages
- System logs: `journalctl -u etl-<pipeline-name>`
- Monitoring logs: `/var/log/etl/monitor.log`

### Monitoring Features
- Automatic log rotation
- Resource usage tracking
- Error statistics
- Email alerts on failure
- Systemd service status

View monitoring information:
```bash
# View pipeline status
./scripts/manage-pipeline.sh info my-pipeline

# Monitor all pipelines
./scripts/monitor-pipelines.sh
```

## Directory Structure

```
etl-framework/
├── cmd/
│   └── etl-cli/          # CLI tool for pipeline management
├── pkg/
│   ├── config/           # Configuration management
│   ├── pipeline/         # Core pipeline interfaces
│   └── env/             # Environment variable handling
├── pipelines/           # Individual ETL pipelines
│   └── my-pipeline/
│       ├── etl.go        # Pipeline implementation
│       ├── config.yaml   # Pipeline configuration
│       └── .env         # Environment variables
├── scripts/            # Management scripts
│   ├── setup-dev.sh    # Development setup
│   ├── manage-pipeline.sh # Pipeline management
│   └── monitor-pipelines.sh # Monitoring utility
└── docker-compose-*.yml # Database configurations
```
  