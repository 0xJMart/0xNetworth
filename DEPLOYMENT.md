# Deployment Guide - YouTube Market Analysis Workflow

## Overview

The workflow feature adds a Python-based agentic service that processes YouTube videos for market analysis. This service runs alongside the existing Go backend and React frontend.

## Architecture

The deployment consists of three services:
1. **Go Backend** - Main API server (port 8080)
2. **Python Workflow Service** - Agentic workflow service (port 8000)
3. **React Frontend** - User interface (port 80)

## Pre-Merge Checklist

### Code Status
- ✅ All Go code compiles successfully
- ✅ Python service structure complete
- ✅ Helm chart templates created
- ✅ Dockerfile for workflow service created
- ✅ Backend configured to connect to workflow service

### What's Ready
- ✅ Core functionality implemented and tested
- ✅ API endpoints for workflow execution and results
- ✅ Helm chart templates for workflow service
- ✅ Backend environment variable configuration

### What Needs to be Done

#### 1. CI/CD Pipeline Updates
The CI/CD pipeline (`.github/workflows/ci-cd.yml`) needs to be updated to:
- Build and publish the Python workflow service Docker image
- Tag as: `ghcr.io/0xjmart/0xnetworth/workflow-service:<version>`
- Include in the Helm chart build process

#### 2. Secrets Configuration
Add `OPENAI_API_KEY` to the Kubernetes secret:
```bash
kubectl create secret generic 0xnetworth-secrets \
  --from-literal=OPENAI_API_KEY=your-openai-api-key \
  --namespace=dev \
  --dry-run=client -o yaml | kubeseal -o yaml > 0xnetworth-sealedsecret.yaml
```

Or update existing sealed secret to include the new key.

#### 3. Helm Chart Values
Update the HelmRelease in Homelab to include workflow service configuration:
- Add `workflow` section to values
- Ensure `OPENAI_API_KEY` is in the secret

## Deployment Structure

### Kubernetes Resources

The Helm chart now deploys:
- `networth-backend` - Go backend service
- `networth-frontend` - React frontend
- `networth-workflow` - Python workflow service (NEW)

### Service Communication

```
Frontend → Backend (port 8080)
Backend → Workflow Service (port 8000)
```

The backend automatically discovers the workflow service via Kubernetes service DNS:
- Service name: `networth-workflow`
- Port: `8000`
- URL: `http://networth-workflow:8000`

## Building and Deploying

### Local Development

1. **Start workflow service:**
   ```bash
   cd workflow-service
   pip3 install -r requirements.txt
   export OPENAI_API_KEY=your_key
   python3 main.py
   ```

2. **Start backend:**
   ```bash
   cd backend
   export WORKFLOW_SERVICE_URL=http://localhost:8000
   go run cmd/server/main.go
   ```

3. **Start frontend:**
   ```bash
   cd frontend
   npm run dev
   ```

### Production Deployment

1. **Build Docker images** (via CI/CD):
   - Backend: `ghcr.io/0xjmart/0xnetworth/backend:<version>`
   - Frontend: `ghcr.io/0xjmart/0xnetworth/frontend:<version>`
   - Workflow: `ghcr.io/0xjmart/0xnetworth/workflow-service:<version>` (NEW)

2. **Deploy via Helm:**
   ```bash
   helm upgrade --install 0xnetworth ./helm/0xnetworth \
     --namespace dev \
     --set workflow.image.tag=<version> \
     --set backend.image.tag=<version> \
     --set frontend.image.tag=<version>
   ```

3. **Or via Flux CD** (GitOps):
   - Update HelmRelease in `Homelab/cluster/apps/0xnetworth/`
   - Flux will automatically reconcile

## Environment Variables

### Backend
- `WORKFLOW_SERVICE_URL` - Workflow service URL (defaults to service name in K8s)
- `COINBASE_API_KEY_NAME` - Coinbase API key (existing)
- `COINBASE_API_PRIVATE_KEY` - Coinbase private key (existing)

### Workflow Service
- `OPENAI_API_KEY` - OpenAI API key (REQUIRED)
- `PORT` - HTTP server port (default: 8000)
- `LOG_LEVEL` - Logging level (default: INFO)

## Resource Requirements

### Workflow Service
- **Requests**: 200m CPU, 512Mi memory
- **Limits**: 1000m CPU, 1Gi memory

These are higher than backend/frontend due to LLM processing requirements.

## Testing in Dev Environment

After deployment, test the workflow:

```bash
# Get the service URL (if using port-forward)
kubectl port-forward -n dev svc/networth-backend 8080:8080

# Test workflow execution
curl -X POST http://localhost:8080/api/workflow/execute \
  -H "Content-Type: application/json" \
  -d '{"youtube_url": "https://www.youtube.com/watch?v=VIDEO_ID"}'

# Get results
curl http://localhost:8080/api/workflow/executions/EXECUTION_ID/details
```

## Next Steps After Merge

1. Update CI/CD workflow to build workflow-service image
2. Add OPENAI_API_KEY to Kubernetes secrets
3. Update HelmRelease values in Homelab cluster
4. Test deployment in dev environment
5. Monitor resource usage and adjust limits if needed

