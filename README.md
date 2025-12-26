# 0xNetworth - Investment Tracking Dashboard

A comprehensive investment tracking dashboard that aggregates data from multiple investment platforms (Coinbase, M1 Finance) to provide a unified view of net worth and investment performance.

## Architecture

- **React Frontend** - Modern dashboard displaying investment data, charts, and net worth calculations
- **Go Backend API** - Aggregates data from Coinbase API and Plaid (for M1 Finance) and serves to frontend
- **Kubernetes Deployment** - Helm chart for easy deployment in homelab cluster

## Project Structure

```
0xNetworth/
├── frontend/          # React application (TypeScript + Vite + Tailwind CSS)
├── backend/           # Go API server (Gin framework)
├── helm/              # Helm chart for Kubernetes deployment
├── scripts/           # Version extraction and utility scripts
└── .github/workflows/ # CI/CD pipeline
```

## Features

### Phase 1: Core Dashboard (In Progress)
- Project initialization and structure
- Backend API foundation
- Frontend dashboard foundation

### Phase 2: Coinbase Integration (Planned)
- Coinbase API integration
- Account balance tracking
- Portfolio data visualization

### Phase 3: M1 Finance Integration (Planned)
- Plaid API integration for M1 Finance
- Investment holdings tracking
- Account aggregation

### Phase 4: Dashboard & Visualization (Planned)
- Net worth calculation and display
- Investment performance charts
- Platform-specific summaries

### Phase 5: Deployment (Planned)
- Helm chart for Kubernetes
- CI/CD pipeline
- Homelab integration via Flux CD

### Future Enhancements
- Agentic workflow for market analysis
- Historical performance tracking
- Transaction history
- Tax reporting
- Multi-currency support
- Alert system

## Development

### Prerequisites

- **Go 1.21+** - [Install Go](https://golang.org/doc/install)
- **Node.js 18+** - [Install Node.js](https://nodejs.org/)
- **Docker** - For containerization
- **Kubernetes** - For deployment (optional for local development)

### Backend Setup

1. Navigate to the backend directory:
```bash
cd backend
```

2. Install dependencies:
```bash
go mod download
```

3. Run the server:
```bash
go run cmd/server/main.go
```

The backend will start on `http://localhost:8080`.

### Frontend Setup

1. Navigate to the frontend directory:
```bash
cd frontend
```

2. Install dependencies:
```bash
npm install
```

3. Start the development server:
```bash
npm run dev
```

The frontend will start on `http://localhost:5173` (or another port if 5173 is taken).

## API Integration

### Coinbase

Coinbase integration uses the Coinbase Advanced Trade API or Coinbase Pro API. You'll need to:
1. Create an API key in your Coinbase account
2. Set environment variables: `COINBASE_API_KEY` and `COINBASE_API_SECRET`

### M1 Finance (via Plaid)

M1 Finance integration uses Plaid API since M1 Finance doesn't provide a public API. You'll need to:
1. Sign up for a Plaid account
2. Create a Plaid application
3. Set environment variables: `PLAID_CLIENT_ID`, `PLAID_SECRET`, and `PLAID_ENVIRONMENT`

## Deployment

The project includes a Helm chart for Kubernetes deployment and a CI/CD pipeline that publishes Docker images and Helm charts to GitHub Container Registry (GHCR).

### Using Published Helm Chart

```bash
# Add the OCI registry (if needed)
helm registry login ghcr.io

# Install from OCI registry
helm install 0xnetworth oci://ghcr.io/0xjmart/0xnetworth/0xnetworth --version 1.0.0
```

## Security

- Never commit API keys to the repository
- Use Kubernetes secrets for sensitive data
- All API keys should be stored as environment variables or secrets
- HTTPS only in production

## License

MIT

