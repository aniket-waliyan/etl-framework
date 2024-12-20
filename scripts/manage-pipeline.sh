#!/bin/bash

# Exit on any error
set -e

# Function to show usage
show_usage() {
    echo "Usage: $0 <command> <pipeline_name> [options]"
    echo "Commands:"
    echo "  start    - Start a pipeline"
    echo "  stop     - Stop a pipeline"
    echo "  status   - Check pipeline status"
    echo "  logs     - View pipeline logs"
    echo "  list     - List all pipelines"
    echo "  enable   - Enable pipeline to start on boot"
    echo "  disable  - Disable pipeline from starting on boot"
    echo "  delete   - Delete pipeline service and logs"
    echo "  info     - Show detailed pipeline information"
    echo "  schedule - Configure pipeline schedule"
    echo
    echo "Schedule Examples:"
    echo "  $0 schedule my-pipeline daily 00:00      # Run daily at midnight"
    echo "  $0 schedule my-pipeline hourly 2         # Run every 2 hours"
    echo "  $0 schedule my-pipeline weekly mon 00:00  # Run weekly on Monday at midnight"
    echo "  $0 schedule my-pipeline monthly 1 00:00   # Run monthly on 1st at midnight"
    echo "  $0 schedule my-pipeline remove           # Remove scheduling"
}

# Function to check if pipeline exists
check_pipeline() {
    local pipeline_name=$1
    if [ ! -d "pipelines/$pipeline_name" ]; then
        echo "Error: Pipeline '$pipeline_name' not found"
        exit 1
    fi
}

# Function to check if service exists
check_service() {
    local pipeline_name=$1
    if [ ! -f "/etc/systemd/system/etl-${pipeline_name}.service" ]; then
        echo "Error: Service for pipeline '$pipeline_name' not found"
        echo "Run setup-dev.sh --prod to create the service"
        exit 1
    fi
}

# Function to get service status with color
get_service_status() {
    local name=$1
    local status=$(systemctl is-active "etl-${name}.service" 2>/dev/null || echo "not configured")
    case $status in
        "active")
            echo -e "\033[32m$status\033[0m"  # Green
            ;;
        "inactive")
            echo -e "\033[31m$status\033[0m"  # Red
            ;;
        *)
            echo -e "\033[33m$status\033[0m"  # Yellow
            ;;
    esac
}

# Function to parse cron expression and show next runs
show_next_runs() {
    local schedule=$1
    local count=5
    
    if command -v cronplan &> /dev/null; then
        echo "Next $count runs:"
        cronplan "$schedule" -n "$count" | while read -r next_run; do
            echo "  - $next_run"
        done
    else
        echo "Install cronplan for schedule prediction: go install github.com/winebarrel/cronplan@latest"
    fi
}

# Function to get last run time
get_last_run() {
    local pipeline_name=$1
    local log_file="/var/log/etl/${pipeline_name}/output.log"
    
    if [ -f "$log_file" ]; then
        local last_run=$(head -n 1 "$log_file" | grep -oE '[0-9]{4}-[0-9]{2}-[0-9]{2} [0-9]{2}:[0-9]{2}:[0-9]{2}' || echo "Never")
        echo "$last_run"
    else
        echo "Never"
    fi
}

# Function to show pipeline configuration
show_pipeline_config() {
    local pipeline_name=$1
    local config_file="pipelines/${pipeline_name}/config.yaml"
    local env_file="pipelines/${pipeline_name}/.env"
    
    echo "Configuration Details:"
    echo "---------------------"
    
    if [ -f "$config_file" ]; then
        echo "Pipeline Configuration (config.yaml):"
        echo "-----------------------------------"
        cat "$config_file"
        echo
    fi
    
    if [ -f "$env_file" ]; then
        echo "Environment Variables (.env):"
        echo "---------------------------"
        # Show env vars but mask sensitive values
        grep -v '^#' "$env_file" | sed 's/=.*$/=****/'
        echo
    fi
    
    # Show systemd service details if exists
    local service_file="/etc/systemd/system/etl-${pipeline_name}.service"
    if [ -f "$service_file" ]; then
        echo "Service Configuration:"
        echo "--------------------"
        systemctl show "etl-${pipeline_name}.service" --property=Description,ExecStart,WorkingDirectory,User,Restart,RestartSec
        echo
    fi
}

# Function to create systemd timer
create_timer() {
    local pipeline_name=$1
    local schedule_type=$2
    local time_spec=$3
    local extra_spec=$4
    local timer_path="/etc/systemd/system/etl-${pipeline_name}.timer"
    local service_name="etl-${pipeline_name}.service"
    local timer_content=""
    
    case $schedule_type in
        "daily")
            # Parse HH:MM
            local hour=${time_spec%:*}
            local minute=${time_spec#*:}
            timer_content="[Unit]
Description=Timer for ETL Pipeline ${pipeline_name} (Daily at ${time_spec})
[Timer]
OnCalendar=*-*-* ${hour}:${minute}:00
Persistent=true
[Install]
WantedBy=timers.target"
            ;;
            
        "hourly")
            # time_spec is the interval in hours
            timer_content="[Unit]
Description=Timer for ETL Pipeline ${pipeline_name} (Every ${time_spec} hours)
[Timer]
OnCalendar=*:0/$(( time_spec * 60 ))
Persistent=true
[Install]
WantedBy=timers.target"
            ;;
            
        "weekly")
            # Parse day and time
            local day=$time_spec
            local time=$extra_spec
            local hour=${time%:*}
            local minute=${time#*:}
            timer_content="[Unit]
Description=Timer for ETL Pipeline ${pipeline_name} (Weekly on ${day} at ${time})
[Timer]
OnCalendar=${day} *-*-* ${hour}:${minute}:00
Persistent=true
[Install]
WantedBy=timers.target"
            ;;
            
        "monthly")
            # Parse day and time
            local day=$time_spec
            local time=$extra_spec
            local hour=${time%:*}
            local minute=${time#*:}
            timer_content="[Unit]
Description=Timer for ETL Pipeline ${pipeline_name} (Monthly on ${day} at ${time})
[Timer]
OnCalendar=*-*-${day} ${hour}:${minute}:00
Persistent=true
[Install]
WantedBy=timers.target"
            ;;
            
        *)
            echo "Error: Invalid schedule type"
            show_usage
            exit 1
            ;;
    esac
    
    # Write timer file
    echo "$timer_content" | sudo tee "$timer_path" > /dev/null
    
    # Reload systemd and enable timer
    sudo systemctl daemon-reload
    sudo systemctl enable "etl-${pipeline_name}.timer"
    sudo systemctl start "etl-${pipeline_name}.timer"
    
    echo "Timer created and enabled successfully"
    echo "Next runs:"
    systemctl list-timers "etl-${pipeline_name}.timer" --no-pager
}

# Parse command line arguments
if [ $# -lt 1 ]; then
    show_usage
    exit 1
fi

command=$1
pipeline_name=$2

case $command in
    "list")
        echo "Scanning for ETL pipelines..."
        echo "----------------------------"
        
        # Check if pipelines directory exists
        if [ ! -d "pipelines" ]; then
            echo "No pipelines directory found!"
            exit 1
        fi
        
        # Count available pipelines
        pipeline_count=0
        
        for pipeline_dir in pipelines/*/; do
            if [ -d "$pipeline_dir" ] && [ -f "${pipeline_dir}/etl.go" ]; then
                name=$(basename "$pipeline_dir")
                status=$(get_service_status "$name")
                
                # Get additional info
                config_file="${pipeline_dir}/config.yaml"
                description=""
                if [ -f "$config_file" ]; then
                    description=$(grep "description:" "$config_file" | cut -d'"' -f2)
                fi
                
                # Get memory usage if service is active
                memory_usage=""
                if [ "$(systemctl is-active "etl-${name}.service" 2>/dev/null)" = "active" ]; then
                    pid=$(systemctl show -p MainPID -v "etl-${name}.service" | cut -d= -f2)
                    if [ -n "$pid" ] && [ "$pid" != "0" ]; then
                        memory_usage=$(ps -o %mem,rss -p "$pid" --no-headers 2>/dev/null || echo "N/A")
                    fi
                fi
                
                echo "Pipeline: $name"
                echo "Status: $status"
                [ -n "$description" ] && echo "Description: $description"
                [ -n "$memory_usage" ] && echo "Memory Usage: $memory_usage"
                echo "----------------------------"
                
                ((pipeline_count++))
            fi
        done
        
        if [ $pipeline_count -eq 0 ]; then
            echo "No ETL pipelines found!"
            echo "Use 'etl-cli generate --name <pipeline_name>' to create a new pipeline."
            exit 0
        fi
        
        echo "Total pipelines found: $pipeline_count"
        ;;
        
    "start")
        if [ -z "$pipeline_name" ]; then
            echo "Error: Pipeline name required"
            show_usage
            exit 1
        fi
        check_pipeline "$pipeline_name"
        check_service "$pipeline_name"
        echo "Starting pipeline ${pipeline_name}..."
        sudo systemctl start "etl-${pipeline_name}.service"
        ;;
        
    "stop")
        if [ -z "$pipeline_name" ]; then
            echo "Error: Pipeline name required"
            show_usage
            exit 1
        fi
        check_pipeline "$pipeline_name"
        check_service "$pipeline_name"
        echo "Stopping pipeline ${pipeline_name}..."
        sudo systemctl stop "etl-${pipeline_name}.service"
        ;;
        
    "status")
        if [ -z "$pipeline_name" ]; then
            echo "Error: Pipeline name required"
            show_usage
            exit 1
        fi
        check_pipeline "$pipeline_name"
        check_service "$pipeline_name"
        sudo systemctl status "etl-${pipeline_name}.service"
        ;;
        
    "logs")
        if [ -z "$pipeline_name" ]; then
            echo "Error: Pipeline name required"
            show_usage
            exit 1
        fi
        check_pipeline "$pipeline_name"
        check_service "$pipeline_name"
        echo "Output logs:"
        echo "------------"
        tail -f "/var/log/etl/${pipeline_name}/output.log" &
        echo "Error logs:"
        echo "-----------"
        tail -f "/var/log/etl/${pipeline_name}/error.log"
        ;;
        
    "enable")
        if [ -z "$pipeline_name" ]; then
            echo "Error: Pipeline name required"
            show_usage
            exit 1
        fi
        check_pipeline "$pipeline_name"
        check_service "$pipeline_name"
        echo "Enabling pipeline ${pipeline_name} to start on boot..."
        sudo systemctl enable "etl-${pipeline_name}.service"
        ;;
        
    "disable")
        if [ -z "$pipeline_name" ]; then
            echo "Error: Pipeline name required"
            show_usage
            exit 1
        fi
        check_pipeline "$pipeline_name"
        check_service "$pipeline_name"
        echo "Disabling pipeline ${pipeline_name} from starting on boot..."
        sudo systemctl disable "etl-${pipeline_name}.service"
        ;;
        
    "delete")
        if [ -z "$pipeline_name" ]; then
            echo "Error: Pipeline name required"
            show_usage
            exit 1
        fi
        check_pipeline "$pipeline_name"
        
        # Confirm deletion
        read -p "Are you sure you want to delete the service and logs for $pipeline_name? [y/N] " confirm
        if [[ $confirm != [yY] ]]; then
            echo "Operation cancelled"
            exit 0
        fi
        
        # Stop and disable the service if it exists
        if [ -f "/etc/systemd/system/etl-${pipeline_name}.service" ]; then
            echo "Stopping and disabling service..."
            sudo systemctl stop "etl-${pipeline_name}.service" 2>/dev/null || true
            sudo systemctl disable "etl-${pipeline_name}.service" 2>/dev/null || true
            
            # Remove service file
            echo "Removing service file..."
            sudo rm -f "/etc/systemd/system/etl-${pipeline_name}.service"
            sudo systemctl daemon-reload
        fi
        
        # Remove logs
        if [ -d "/var/log/etl/${pipeline_name}" ]; then
            echo "Removing log files..."
            sudo rm -rf "/var/log/etl/${pipeline_name}"
        fi
        
        # Remove logrotate config
        if [ -f "/etc/logrotate.d/etl-${pipeline_name}" ]; then
            echo "Removing logrotate configuration..."
            sudo rm -f "/etc/logrotate.d/etl-${pipeline_name}"
        fi
        
        echo "Pipeline service and logs deleted successfully"
        ;;
        
    "info")
        if [ -z "$pipeline_name" ]; then
            echo "Error: Pipeline name required"
            show_usage
            exit 1
        fi
        check_pipeline "$pipeline_name"
        
        echo "Pipeline Service Information: $pipeline_name"
        echo "====================================="
        
        # Show systemd service status and configuration
        if [ -f "/etc/systemd/system/etl-${pipeline_name}.service" ]; then
            echo "Service Status:"
            echo "--------------"
            systemctl status "etl-${pipeline_name}.service" --no-pager
            
            echo -e "\nService Configuration:"
            echo "--------------------"
            # Show full systemd service configuration
            systemctl cat "etl-${pipeline_name}.service"
            
            echo -e "\nService Properties:"
            echo "-----------------"
            # Show important service properties
            systemctl show "etl-${pipeline_name}.service" \
                --property=Id \
                --property=Description \
                --property=LoadState \
                --property=ActiveState \
                --property=SubState \
                --property=UnitFileState \
                --property=ExecStart \
                --property=ExecMainPID \
                --property=MainPID \
                --property=Type \
                --property=Restart \
                --property=RestartSec \
                --property=Environment \
                --property=EnvironmentFile \
                --property=WorkingDirectory \
                --property=WatchdogTimestamp \
                --property=WatchdogTimestampMonotonic
            
            echo -e "\nTimer Information:"
            echo "-----------------"
            # Show timer details if exists
            if [ -f "/etc/systemd/system/etl-${pipeline_name}.timer" ]; then
                systemctl cat "etl-${pipeline_name}.timer"
                echo -e "\nTimer Status:"
                systemctl status "etl-${pipeline_name}.timer" --no-pager
                echo -e "\nNext Scheduled Runs:"
                systemctl list-timers "etl-${pipeline_name}.timer" --no-pager
            else
                echo "No timer configured. Service runs continuously."
            fi
            
            echo -e "\nResource Usage:"
            echo "--------------"
            pid=$(systemctl show -p MainPID -v "etl-${pipeline_name}.service" | cut -d= -f2)
            if [ -n "$pid" ] && [ "$pid" != "0" ]; then
                ps -p "$pid" -o pid,ppid,user,%cpu,%mem,vsz,rss,stat,start,time,cmd --headers
            else
                echo "Service not running"
            fi
            
            echo -e "\nJournal Logs (last 5 entries):"
            echo "-----------------------------"
            journalctl -u "etl-${pipeline_name}.service" -n 5 --no-pager
        else
            echo "Service not installed. Run setup-dev.sh --prod to create the service."
        fi
        ;;
        
    "schedule")
        if [ -z "$pipeline_name" ]; then
            echo "Error: Pipeline name required"
            show_usage
            exit 1
        fi
        check_pipeline "$pipeline_name"
        check_service "$pipeline_name"
        
        schedule_type=$3
        time_spec=$4
        extra_spec=$5
        
        if [ "$schedule_type" = "remove" ]; then
            # Remove existing timer
            if [ -f "/etc/systemd/system/etl-${pipeline_name}.timer" ]; then
                echo "Removing timer for ${pipeline_name}..."
                sudo systemctl stop "etl-${pipeline_name}.timer"
                sudo systemctl disable "etl-${pipeline_name}.timer"
                sudo rm -f "/etc/systemd/system/etl-${pipeline_name}.timer"
                sudo systemctl daemon-reload
                echo "Timer removed successfully"
            else
                echo "No timer found for ${pipeline_name}"
            fi
        else
            if [ -z "$schedule_type" ] || [ -z "$time_spec" ]; then
                echo "Error: Schedule type and time required"
                show_usage
                exit 1
            fi
            
            # Create new timer
            create_timer "$pipeline_name" "$schedule_type" "$time_spec" "$extra_spec"
        fi
        ;;
        
    *)
        echo "Error: Unknown command '$command'"
        show_usage
        exit 1
        ;;
esac 