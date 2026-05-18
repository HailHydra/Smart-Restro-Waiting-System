# Frontend

Svelte dashboard for the SmartRestro queue system.

## Environment

Create `.env` from `.env.example` and set:

```bash
VITE_API_BASE_URL=http://localhost:8080/api/v1
```

## Commands

```bash
npm install
npm run dev
npm run build
```

## Current capabilities

- loads restaurants from the Go API
- shows queue summary and queue entries
- adds new queue entries
- advances queue state transitions
- reads the API base URL from Vite env
