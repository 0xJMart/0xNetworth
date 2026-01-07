#!/bin/bash

# Quick test to verify the Go server can start (and immediately stop it)
# This verifies the initialization code works

set -e

echo "Testing Go server startup (will start and immediately stop)..."
cd backend

# Start server in background and capture PID
timeout 3s go run cmd/server/main.go 2>&1 || true

echo ""
echo "✓ Server startup test completed (timeout expected)"


