package env

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"
)

// Config holds environment variables
type Config struct {
	// SQL Server
	SQLServerUser     string
	SQLServerPassword string
	SQLServerShards   []string // List of shard connection strings
	SQLServerDB       string

	// PostgreSQL
	PostgresUser     string
	PostgresPassword string
	PostgresDB       string
	PostgresHost     string
	PostgresPort     string

	// ETL
	RetryCount string
	RetryDelay string
}

// Load reads environment variables from .env file
func Load(workDir string) (*Config, error) {
	envFile := filepath.Join(workDir, ".env")
	if err := godotenv.Load(envFile); err != nil {
		return nil, fmt.Errorf("error loading .env file: %v", err)
	}

	// Parse shard configuration
	shardHosts := getEnvOrDefault("SQLSERVER_SHARD_HOSTS", "127.0.0.1")
	shardPorts := getEnvOrDefault("SQLSERVER_SHARD_PORTS", "1433,1434,1435,1436")

	hosts := strings.Split(shardHosts, ",")
	ports := strings.Split(shardPorts, ",")

	var shards []string
	for i, port := range ports {
		host := "127.0.0.1" // default host
		if i < len(hosts) {
			host = strings.TrimSpace(hosts[i])
		}
		shards = append(shards, fmt.Sprintf("%s:%s", host, strings.TrimSpace(port)))
	}

	return &Config{
		// SQL Server
		SQLServerUser:     getEnvOrDefault("SQLSERVER_USER", "sa"),
		SQLServerPassword: getEnvOrDefault("SQLSERVER_PASSWORD", ""),
		SQLServerShards:   shards,
		SQLServerDB:       getEnvOrDefault("SQLSERVER_DB", "my_source_db"),

		// PostgreSQL
		PostgresUser:     getEnvOrDefault("POSTGRES_USER", "etl_user"),
		PostgresPassword: getEnvOrDefault("POSTGRES_PASSWORD", ""),
		PostgresDB:       getEnvOrDefault("POSTGRES_DB", "my_sink_db"),
		PostgresHost:     getEnvOrDefault("POSTGRES_HOST", "127.0.0.1"),
		PostgresPort:     getEnvOrDefault("POSTGRES_PORT", "5432"),

		// ETL
		RetryCount: getEnvOrDefault("ETL_RETRY_COUNT", "3"),
		RetryDelay: getEnvOrDefault("ETL_RETRY_DELAY", "5s"),
	}, nil
}

// getEnvOrDefault returns environment variable value or default if not set
func getEnvOrDefault(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
