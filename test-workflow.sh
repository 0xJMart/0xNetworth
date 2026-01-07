#!/bin/bash

# Test script for workflow implementation
set -e

echo "=== Testing 0xNetworth Workflow Implementation ==="
echo ""

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test 1: Go backend compilation
echo "1. Testing Go backend compilation..."
cd backend
if go build ./cmd/server/main.go 2>&1; then
    echo -e "${GREEN}✓ Go backend compiles successfully${NC}"
    rm -f main
else
    echo -e "${RED}✗ Go backend compilation failed${NC}"
    exit 1
fi
cd ..

# Test 2: Python service imports (without installing dependencies)
echo ""
echo "2. Testing Python service structure..."
cd workflow-service

# Check if main files exist
if [ -f "main.py" ] && [ -f "models.py" ] && [ -f "requirements.txt" ]; then
    echo -e "${GREEN}✓ Python service files exist${NC}"
else
    echo -e "${RED}✗ Python service files missing${NC}"
    exit 1
fi

# Check if agent files exist
if [ -f "agents/transcript_agent.py" ] && [ -f "agents/analysis_agent.py" ] && [ -f "agents/recommendation_agent.py" ]; then
    echo -e "${GREEN}✓ Agent files exist${NC}"
else
    echo -e "${RED}✗ Agent files missing${NC}"
    exit 1
fi

# Check if tool files exist
if [ -f "tools/youtube_tool.py" ]; then
    echo -e "${GREEN}✓ Tool files exist${NC}"
else
    echo -e "${RED}✗ Tool files missing${NC}"
    exit 1
fi

cd ..

# Test 3: Check Go models compile
echo ""
echo "3. Testing Go models..."
cd backend
if go build ./internal/models/... 2>&1; then
    echo -e "${GREEN}✓ Go models compile successfully${NC}"
else
    echo -e "${RED}✗ Go models compilation failed${NC}"
    exit 1
fi
cd ..

# Test 4: Check workflow package compiles
echo ""
echo "4. Testing workflow package..."
cd backend
if go build ./internal/workflow/... 2>&1; then
    echo -e "${GREEN}✓ Workflow package compiles successfully${NC}"
else
    echo -e "${RED}✗ Workflow package compilation failed${NC}"
    exit 1
fi
cd ..

# Test 5: Check handlers compile
echo ""
echo "5. Testing handlers package..."
cd backend
if go build ./internal/handlers/... 2>&1; then
    echo -e "${GREEN}✓ Handlers package compiles successfully${NC}"
else
    echo -e "${RED}✗ Handlers package compilation failed${NC}"
    exit 1
fi
cd ..

# Test 6: Check integrations compile
echo ""
echo "6. Testing integrations package..."
cd backend
if go build ./internal/integrations/workflow/... 2>&1; then
    echo -e "${GREEN}✓ Workflow integration package compiles successfully${NC}"
else
    echo -e "${RED}✗ Workflow integration package compilation failed${NC}"
    exit 1
fi
cd ..

# Test 7: Check store compiles
echo ""
echo "7. Testing store package..."
cd backend
if go build ./internal/store/... 2>&1; then
    echo -e "${GREEN}✓ Store package compiles successfully${NC}"
else
    echo -e "${RED}✗ Store package compilation failed${NC}"
    exit 1
fi
cd ..

echo ""
echo -e "${GREEN}=== All basic tests passed! ===${NC}"
echo ""
echo -e "${YELLOW}Note:${NC} For full testing, you'll need to:"
echo "  1. Install Python dependencies: cd workflow-service && pip3 install -r requirements.txt"
echo "  2. Set OPENAI_API_KEY environment variable for the Python service"
echo "  3. Run the Python service: cd workflow-service && python3 main.py"
echo "  4. Run the Go backend: cd backend && go run cmd/server/main.go"
echo "  5. Test the /api/workflow/execute endpoint with a YouTube URL"


