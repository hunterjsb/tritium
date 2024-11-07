#!/bin/bash
# stop-cluster.sh

echo "Stopping Tritium cluster..."

# Stop Tritium nodes
for pidfile in logs/node*.pid; do
    if [ -f "$pidfile" ]; then
        pid=$(cat "$pidfile")
        echo "Stopping node with PID $pid"
        kill $pid 2>/dev/null || true
        rm "$pidfile"
    fi
done

# Stop Redis instances
echo "Stopping Redis instances..."
for port in 6379 6380 6381 6382 6383 6384; do
    redis-cli -p $port shutdown || true
done

echo "Cleanup complete!"