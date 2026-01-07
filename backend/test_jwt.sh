#!/bin/bash

# Test script for JWT generation
# This script generates a JWT and tests it with curl

set -e

echo "=== Testing JWT Generation ==="
echo ""

# Check if environment variables are set
if [ -z "$COINBASE_API_KEY_NAME" ] || [ -z "$COINBASE_API_PRIVATE_KEY" ]; then
    echo "Error: COINBASE_API_KEY_NAME and COINBASE_API_PRIVATE_KEY must be set"
    echo ""
    echo "Example:"
    echo "  export COINBASE_API_KEY_NAME='your-api-key-id'"
    echo "  export COINBASE_API_PRIVATE_KEY='your-private-key'"
    exit 1
fi

# Build and run the test program
echo "Generating JWT..."
cd "$(dirname "$0")"
go run test_jwt.go

echo ""
echo "=== Manual Test with curl ==="
echo ""
echo "After running the test program above, you can manually test with:"
echo ""
echo "export JWT='<jwt-from-output-above>'"
echo ""
echo "curl -L -X GET 'https://api.coinbase.com/api/v3/brokerage/accounts' \\"
echo "  -H \"Authorization: Bearer \$JWT\" \\"
echo "  -H \"Content-Type: application/json\" \\"
echo "  -H \"Accept: application/json\""
echo ""




