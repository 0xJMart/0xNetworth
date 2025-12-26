# 0xNetworth Backend

Go API server for the 0xNetworth investment tracking dashboard.

## Structure

```
backend/
├── cmd/
│   └── server/          # Main server application
│       └── main.go
├── internal/
│   ├── models/         # Data models (Account, Investment, NetWorth, Transaction)
│   ├── handlers/        # HTTP request handlers
│   ├── store/          # In-memory data store
│   └── integrations/   # External API clients
│       └── coinbase/   # Coinbase API client (Phase 4)
└── go.mod
```

## Development

### Prerequisites

- Go 1.21+

### Setup

1. Install dependencies:
```bash
go mod download
```

2. Run the server:
```bash
go run cmd/server/main.go
```

The server will start on `http://localhost:8080` (or the port specified in `PORT` environment variable).

### Environment Variables

Create a `.env` file (or set environment variables):

```bash
PORT=8080

# Coinbase API (Phase 4)
COINBASE_API_KEY=your_api_key
COINBASE_API_SECRET=your_api_secret

```

## API Endpoints

### Health Check
- `GET /api/health` - Health check endpoint

### Accounts
- `GET /api/accounts` - Get all accounts
- `GET /api/accounts/platform/:platform` - Get accounts by platform (coinbase)
- `GET /api/accounts/:id` - Get account by ID

### Investments
- `GET /api/investments` - Get all investments
- `GET /api/investments/account/:accountId` - Get investments by account ID
- `GET /api/investments/platform/:platform` - Get investments by platform

### Net Worth
- `GET /api/networth` - Get current net worth
- `GET /api/networth/breakdown` - Get detailed net worth breakdown

### Sync
- `POST /api/sync` - Trigger sync from all platforms
- `POST /api/sync/:platform` - Trigger sync for specific platform

## Current Status

- ✅ Phase 2: Backend foundation complete
- ⏳ Phase 4: Coinbase integration (placeholder ready)

