#!/bin/bash
# start-cluster.sh

# Function to wait for Redis to be ready
wait_for_redis() {
    local port=$1
    local max_attempts=30
    local attempt=1

    echo -n "Waiting for Redis on port $port "
    while ! redis-cli -p $port ping >/dev/null 2>&1; do
        if [ $attempt -ge $max_attempts ]; then
            echo "Failed to connect to Redis on port $port after $max_attempts attempts"
            return 1
        fi
        echo -n "."
        sleep 1
        attempt=$((attempt + 1))
    done
    echo " ready!"
    return 0
}

# Create required directories
mkdir -p data/{primary,replica1,replica2} logs

echo "Starting Redis instances..."

# Start primary nodes
redis-server --port 6379 --dir ./data/primary --daemonize yes
redis-server --port 6381 --dir ./data/replica1 --daemonize yes
redis-server --port 6383 --dir ./data/replica2 --daemonize yes

# Wait for primary nodes
wait_for_redis 6379
wait_for_redis 6381
wait_for_redis 6383

# Start replica nodes
redis-server --port 6380 --dir ./data/primary --replicaof localhost 6379 --daemonize yes
redis-server --port 6382 --dir ./data/replica1 --replicaof localhost 6381 --daemonize yes
redis-server --port 6384 --dir ./data/replica2 --replicaof localhost 6383 --daemonize yes

# Wait for replica nodes
wait_for_redis 6380
wait_for_redis 6382
wait_for_redis 6384

echo "Verifying replication status..."
for port in 6380 6382 6384; do
    master_port=$((port - 1))
    echo -n "Checking replica on port $port -> master on $master_port: "
    
    # Wait for replication to be established
    max_attempts=30
    attempt=1
    while true; do
        status=$(redis-cli -p $port info replication | grep master_link_status)
        if [[ $status == *"up"* ]]; then
            echo "OK"
            break
        fi
        if [ $attempt -ge $max_attempts ]; then
            echo "WARNING: Replication not established"
            break
        fi
        echo -n "."
        sleep 1
        attempt=$((attempt + 1))
    done
done

echo "Starting Tritium nodes..."

# Start first node and wait for it to be ready
go run cmd/server/main.go -config node1.env > logs/node1.log 2>&1 &
pid1=$!
echo "Node 1 started (PID: $pid1)"

# Wait a bit longer for the first node as it establishes the cluster
sleep 3

# Start remaining nodes
go run cmd/server/main.go -config node2.env > logs/node2.log 2>&1 &
pid2=$!
echo "Node 2 started (PID: $pid2)"

go run cmd/server/main.go -config node3.env > logs/node3.log 2>&1 &
pid3=$!
echo "Node 3 started (PID: $pid3)"

# Wait for Tritium nodes to initialize
sleep 2

echo "Cluster startup complete!"
echo "To stop the cluster, run: ./stop-cluster.sh"

# Store PIDs for the stop script
echo "$pid1" > logs/node1.pid
echo "$pid2" > logs/node2.pid
echo "$pid3" > logs/node3.pid

# Start the monitor
# echo "Starting monitor..."
# exec go run cmd/monitor/main.go