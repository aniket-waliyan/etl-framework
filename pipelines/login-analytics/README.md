# Login-Analytics Pipeline

This pipeline is generated using the ETL Pipeline Framework.

## Configuration

Edit the config.yaml file to configure:
- Source connection details
- Transformation rules
- Sink connection details

## Running the Pipeline

```bash
go run etl.go
```

Or using Docker:

```bash
docker build -t login-analytics .
docker run login-analytics
```
