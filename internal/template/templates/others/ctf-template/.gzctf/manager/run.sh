#!/bin/bash
set -e

cd $WORKDIR

# Function to handle shutdown signals
cleanup() {
    echo "Received shutdown signal"
    
    # Kill the background process (gzcli serve)
    if [ -n "$SERVE_PID" ]; then
        echo "Forwarding signal to gzcli serve (PID: $SERVE_PID)"
        kill -TERM "$SERVE_PID"
        wait "$SERVE_PID"
    fi
    
    # Kill other background processes if necessary
    # (Optional: kill sync, bot, etc. if we tracked them)
    
    exit 0
}

# Trap SIGINT and SIGTERM
trap cleanup SIGINT SIGTERM

echo "Starting GZCLI services..."

# Start background services
echo "Starting sync..."
gzcli sync &

echo "Starting script manager..."
gzcli script start &

echo "Starting log watcher..."
gzcli watch start &

# Start main server loop
echo "Starting server..."
while true; do
    gzcli serve --port 3000 --host 0.0.0.0 &
    SERVE_PID=$!
    
    # Wait for the specific process
    wait "$SERVE_PID"
    
    # If wait exited because of the trap, we are done
    # If wait exited because the process crashed, we restart (loop continues)
    EXIT_CODE=$?
    
    # If the process exited normally or with error (not killed by us), log it
    echo "gzcli serve exited with code $EXIT_CODE. Restarting in 1s..."
    sleep 1
done
