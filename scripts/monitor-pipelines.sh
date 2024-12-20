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
