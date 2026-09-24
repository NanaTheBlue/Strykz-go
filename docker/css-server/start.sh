#!/bin/bash
set -e

echo "Starting CS:S Dedicated Server..."
./srcds_run -game cstrike -console -usercon +map de_dust2 +maxplayers 16 -condebug &
SRCDS_PID=$!

# Ensure the log file exists before the sidecar tries to read it
sleep 2
touch cstrike/console.log

echo "Starting Sidecar..."
# The sidecar picks up SERVER_ID and ORCHESTRATOR_ADDR from the environment
./sidecar -log-file "cstrike/console.log" &
SIDECAR_PID=$!

# Wait for both processes
wait $SRCDS_PID
wait $SIDECAR_PID
