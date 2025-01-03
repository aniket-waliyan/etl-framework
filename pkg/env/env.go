package env

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	
	SQLServerShards   []string 
	SQLServerUser     string
	SQLServerPassword string
	SQLServerDB       string

	
	PostgresUser     string
	PostgresPassword string
	PostgresDB       string
	PostgresHost     string
	PostgresPort     string

	
	RetryCount string
	RetryDelay string
}

func Load(workDir string) (*Config, error) {
	envFile := filepath.Join(workDir, ".env")
	if err := godotenv.Load(envFile); err != nil {
		return nil, fmt.Errorf("error loading .env file: %v", err)
	}

	
	shardHosts := getEnvOrDefault("SQLSERVER_SHARD_HOSTS", "localhost")
	shardPorts := getEnvOrDefault("SQLSERVER_SHARD_PORTS", "1433,1434,1435,1436")

	hosts := strings.Split(shardHosts, ",")
	ports := strings.Split(shardPorts, ",")

	var shards []string
	for i, port := range ports {
		host := "localhost" 
		if i < len(hosts) {
			host = strings.TrimSpace(hosts[i])
		}
		shards = append(shards, fmt.Sprintf("%s:%s", host, strings.TrimSpace(port)))
	}

	return &Config{
		
		SQLServerShards:   shards,
		SQLServerUser:     getEnvOrDefault("SQLSERVER_USER", "sa"),
		SQLServerPassword: getEnvOrDefault("SQLSERVER_PASSWORD", ""),
		SQLServerDB:       getEnvOrDefault("SQLSERVER_DB", "NSEBSE"),

		
		PostgresUser:     getEnvOrDefault("POSTGRES_USER", "etl_user"),
		PostgresPassword: getEnvOrDefault("POSTGRES_PASSWORD", ""),
		PostgresDB:       getEnvOrDefault("POSTGRES_DB", "etl_sink_db"),
		PostgresHost:     getEnvOrDefault("POSTGRES_HOST", "localhost"),
		PostgresPort:     getEnvOrDefault("POSTGRES_PORT", "5432"),

		
		RetryCount: getEnvOrDefault("ETL_RETRY_COUNT", "3"),
		RetryDelay: getEnvOrDefault("ETL_RETRY_DELAY", "5s"),
	}, nil
}

func getEnvOrDefault(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
