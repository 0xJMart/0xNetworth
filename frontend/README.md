# 0xNetworth Frontend

React + TypeScript frontend for the 0xNetworth investment tracking dashboard.

## Structure

```
frontend/
├── src/
│   ├── components/      # React components
│   │   ├── NetWorthCard.tsx
│   │   ├── AccountList.tsx
│   │   ├── InvestmentChart.tsx
│   │   ├── PlatformCard.tsx
│   │   └── SyncButton.tsx
│   ├── api.ts          # API client functions
│   ├── types.ts        # TypeScript type definitions
│   ├── App.tsx         # Main application component
│   ├── main.tsx        # Application entry point
│   └── index.css       # Global styles with Tailwind
└── package.json
```

## Development

### Prerequisites

- Node.js 18+
- npm or yarn

### Setup

1. Install dependencies:
```bash
npm install
```

2. Start the development server:
```bash
npm run dev
```

The frontend will start on `http://localhost:5173` (or another port if 5173 is taken).

The Vite dev server is configured to proxy API requests to `http://localhost:8080` (the backend server).

### Environment Variables

Create a `.env` file (optional):

```bash
VITE_API_URL=http://localhost:8080/api
```

If not set, the frontend will use the Vite proxy configuration.

### Build

Build for production:

```bash
npm run build
```

The built files will be in the `dist/` directory.

### Testing

Run tests:

```bash
npm test
```

Run tests with UI:

```bash
npm run test:ui
```

Run tests with coverage:

```bash
npm run test:coverage
```

## Components

### NetWorthCard
Displays total net worth with breakdown by platform.

### AccountList
Lists all investment accounts with balances and platform information.

### InvestmentChart
Pie chart visualization of investment distribution by asset type (using Recharts).

### PlatformCard
Platform-specific summary cards showing accounts, holdings, and total value.

### SyncButton
Button to trigger data synchronization from external APIs.

## Features

- Real-time net worth calculation
- Platform filtering (Coinbase, M1 Finance)
- Investment distribution visualization
- Account management display
- Manual sync functionality
- Responsive design with Tailwind CSS

## Dependencies

- **React 19+** - UI library
- **TypeScript** - Type safety
- **Vite** - Build tool and dev server
- **Tailwind CSS** - Utility-first CSS framework
- **Recharts** - Chart library for React

## Current Status

- ✅ Phase 3: Frontend foundation complete
- ⏳ Phase 4: Coinbase integration (backend ready)
- ⏳ Phase 5: Plaid/M1 Finance integration (backend ready)

