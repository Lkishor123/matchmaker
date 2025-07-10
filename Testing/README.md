# Testing Suite

This directory contains end-to-end tests for the Matchmaker services.

## Running

1. Start the services (e.g. via `docker-compose up` at the repository root).
2. Install Python dependencies and run the tests:

```bash
pip install -r requirements.txt
pytest
```

The tests read service endpoints from environment variables matching the
defaults defined in `internal/config/config.go`:

- `GATEWAY_URL` (default `http://localhost:8080`)
- `AUTH_SERVICE_URL` (default `http://localhost:8081`)
- `CHAT_SERVICE_URL` (default `http://localhost:8082`)
- `MATCH_SERVICE_URL` (default `http://localhost:8083`)
- `USER_SERVICE_URL` (default `http://localhost:8084`)
- `REPORT_SERVICE_URL` (default `http://localhost:8085`)
- `JWT_PRIVATE_KEY` – RSA key used when exercising the gateway
