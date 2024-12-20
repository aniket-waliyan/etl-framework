#!/bin/bash

# Exit on any error
set -e

# Function to install Go tools with error handling
install_go_tools() {
    echo "Installing development tools..."
    
    # Install golangci-lint
    if ! command -v golangci-lint &> /dev/null; then
        echo "Installing golangci-lint..."
        curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.55.2
    else
        echo "golangci-lint already installed"
    fi

    # Install mockgen
    if ! command -v mockgen &> /dev/null; then
        echo "Installing mockgen..."
        go install github.com/golang/mock/mockgen@latest
    else
        echo "mockgen already installed"
    fi
}

# Function to create systemd service
create_systemd_service() {
    local pipeline_name=$1
    local working_dir=$2
    local service_name="etl-${pipeline_name}.service"
    
    echo "Creating systemd service for ${pipeline_name}..."
    
    # Create service file
    sudo tee "/etc/systemd/system/${service_name}" > /dev/null << EOF
[Unit]
Description=ETL Pipeline Service for ${pipeline_name}
After=network.target postgresql.service

[Service]
Type=simple
User=$USER
WorkingDirectory=${working_dir}/pipelines/${pipeline_name}
ExecStart=/usr/local/go/bin/go run etl.go
Restart=always
RestartSec=10
StandardOutput=append:/var/log/etl/${pipeline_name}/output.log
StandardError=append:/var/log/etl/${pipeline_name}/error.log
Environment="PATH=/usr/local/go/bin:/usr/local/bin:/usr/bin:/bin"
Environment="GOPATH=${HOME}/go"
Environment="GOBIN=${HOME}/go/bin"

# Load environment variables from file
EnvironmentFile=${working_dir}/pipelines/${pipeline_name}/.env

[Install]
WantedBy=multi-user.target
EOF

    # Create log directory with appropriate permissions
    sudo mkdir -p "/var/log/etl/${pipeline_name}"
    sudo chown -R $USER:$USER "/var/log/etl/${pipeline_name}"
    sudo chmod 755 "/var/log/etl/${pipeline_name}"

    # Reload systemd and enable service
    sudo systemctl daemon-reload
    sudo systemctl enable "${service_name}"
    echo "Service ${service_name} created and enabled"
}

# Function to create logrotate configuration
create_logrotate_config() {
    local pipeline_name=$1
    
    echo "Creating logrotate configuration for ${pipeline_name}..."
    
    sudo tee "/etc/logrotate.d/etl-${pipeline_name}" > /dev/null << EOF
/var/log/etl/${pipeline_name}/*.log {
    daily
    rotate 7
    compress
    delaycompress
    missingok
    notifempty
    create 0644 $USER $USER
}
EOF
}

# Main setup
echo "Setting up environment..."

# Check if running as root
if [ "$EUID" -eq 0 ]; then 
    echo "Please run without sudo"
    exit 1
fi

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "Error: Go is not installed"
    exit 1
fi

# Install development tools
install_go_tools

# Install dependencies
echo "Installing project dependencies..."
go mod download
go mod tidy

# Create build directory
mkdir -p build

# Setup for production if specified
if [ "$1" == "--prod" ]; then
    echo "Setting up production environment..."
    
    # Get absolute path of the project
    WORKING_DIR=$(pwd)
    
    # Create base log directory
    sudo mkdir -p /var/log/etl
    sudo chown -R $USER:$USER /var/log/etl
    
    # Scan for pipeline directories
    for pipeline_dir in pipelines/*/; do
        if [ -d "$pipeline_dir" ]; then
            pipeline_name=$(basename "$pipeline_dir")
            
            # Create service and logrotate config for each pipeline
            create_systemd_service "$pipeline_name" "$WORKING_DIR"
            create_logrotate_config "$pipeline_name"
            
            # Copy .env.template to .env if it doesn't exist
            if [ ! -f "${pipeline_dir}/.env" ]; then
                cp .env.template "${pipeline_dir}/.env"
                echo "Created .env file for ${pipeline_name}. Please update with appropriate values."
            fi
            
            echo "Pipeline ${pipeline_name} configured for production"
        fi
    done
    
    # Create monitoring script
    cat > scripts/monitor-pipelines.sh << 'EOF'
#!/bin/bash

echo "ETL Pipeline Status Report - $(date)"
echo "================================="

for service in /etc/systemd/system/etl-*.service; do
    name=$(basename "$service")
    status=$(systemctl is-active "$name")
    echo "${name}: ${status}"
    
    if [ "$status" = "active" ]; then
        # Get last 5 lines of logs
        echo "Recent logs:"
        pipeline_name=${name%.service}
        pipeline_name=${pipeline_name#etl-}
        tail -n 5 "/var/log/etl/${pipeline_name}/output.log"
        echo "Memory usage:"
        ps -o pid,ppid,%mem,rss,cmd -p $(systemctl show -p MainPID -v "$name" | cut -d= -f2)
        echo "--------------------------------"
    fi
done
EOF
    
    chmod +x scripts/monitor-pipelines.sh
    
    # Create cron job for monitoring
    (crontab -l 2>/dev/null; echo "0 * * * * $WORKING_DIR/scripts/monitor-pipelines.sh >> /var/log/etl/monitor.log 2>&1") | crontab -
    
    echo "Production setup complete!"
    echo "Use the following commands to manage services:"
    echo "- Start a pipeline: sudo systemctl start etl-<pipeline_name>"
    echo "- Stop a pipeline:  sudo systemctl stop etl-<pipeline_name>"
    echo "- Check status:     sudo systemctl status etl-<pipeline_name>"
    echo "- View logs:        tail -f /var/log/etl/<pipeline_name>/output.log"
else
    echo "Development environment setup complete!"
fi 