#!/bin/bash

# Function to test SQL Server connection
test_sqlserver() {
    local port=$1
    /opt/mssql-tools18/bin/sqlcmd -S localhost,$port -U sa -P 'jdwnjw9de3' -Q "SELECT 1" -C -N > /dev/null 2>&1
    return $?
}

# Function to test PostgreSQL connection
test_postgres() {
    PGPASSWORD=jdwnjw9de3 psql -h localhost -U etl_user -d my_sink_db -c "SELECT 1" > /dev/null 2>&1
    return $?
}

echo "Waiting for databases to be ready..."

# Wait for all databases to be ready (up to 5 minutes)
for i in {1..30}; do
    all_ready=true
    
    # Check SQL Server shards
    for port in {1433..1436}; do
        if ! test_sqlserver $port; then
            all_ready=false
            echo "Waiting for SQL Server on port $port..."
        fi
    done
    
    # Check PostgreSQL
    if ! test_postgres; then
        all_ready=false
        echo "Waiting for PostgreSQL..."
    fi
    
    if $all_ready; then
        break
    fi
    
    sleep 10
done

# Initialize SQL Server shards
for i in {1433..1436}; do
    echo "Initializing SQL Server shard on port $i..."
    if /opt/mssql-tools18/bin/sqlcmd -S localhost,$i -U sa -P 'jdwnjw9de3' -i init-source-db.sql -C -N; then
        echo "Successfully initialized shard on port $i"
    else
        echo "Failed to initialize shard on port $i"
    fi
done

# Initialize PostgreSQL sink database
echo "Initializing PostgreSQL sink database..."
if PGPASSWORD=jdwnjw9de3 psql -h localhost -U etl_user -d my_sink_db -f init-sink-db.sql; then
    echo "Successfully initialized PostgreSQL database"
else
    echo "Failed to initialize PostgreSQL database"
fi

echo "Database initialization completed!" 